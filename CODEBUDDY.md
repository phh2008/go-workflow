# CODEBUDDY.md

This file provides guidance to CodeBuddy Code when working with code in this repository.

## Project Overview

easy-workflow 是一个纯 Go 语言开发的简单工作流引擎，支持作为库集成到 Go 项目中，也可独立作为 Web API Server 运行。专为"中国式流程"设计，支持会签、混合网关、自由驳回、直接跳转等特性。

**技术栈**: Go 1.21+、MySQL 8.0+（CTE）、GORM、Gin（Web API）、expr-lang/expr（表达式求值）、slog（日志）

## Build & Run

```bash
# 构建
go build ./...

# 运行示例（需要 MySQL 连接，配置在 example/main.go 中）
go run example/main.go

# 运行测试
go test ./...

# 运行单个包的测试
go test ./internal/service/...
```

**注意**: Service 层有单元测试（`internal/service/*_test.go`，68 个用例）。

## Architecture

### 项目结构

采用分层架构 + struct-based 设计，所有状态封装在 `Engine` 结构体中，通过构造函数注入依赖。`internal/` 下的包不对外暴露，库的 API 通过根包 `easyworkflow` 提供。

```
easy-workflow/
├── engine.go                          # 公共 API 入口 (Engine, Config, New, StartWebAPI)
├── internal/
│   ├── entity/                        # 数据库实体（GORM 模型）
│   │   ├── local_time.go              # 自定义时间类型
│   │   ├── proc_def.go                # ProcDef, HistProcDef
│   │   ├── proc_inst.go               # ProcInst, HistProcInst
│   │   ├── proc_task.go               # ProcTask, HistProcTask
│   │   ├── proc_execution.go          # ProcExecution, HistProcExecution
│   │   └── proc_inst_variable.go      # ProcInstVariable, HistProcInstVariable
│   ├── model/                         # 领域模型（与 DB 无关的业务结构）
│   │   ├── node.go                    # NodeType 常量, Node 结构体
│   │   ├── process.go                 # Process 结构体
│   │   ├── gateway.go                 # HybridGateway, Condition
│   │   ├── task_view.go               # TaskView（任务查询视图模型）
│   │   ├── instance_view.go           # InstanceView（实例查询视图模型）
│   │   ├── action.go                  # TaskAction
│   │   ├── variable.go                # Variable
│   │   └── request.go                 # 请求/响应参数结构体
│   ├── repository/                    # 数据访问层
│   │   ├── interfaces.go              # Repository 接口定义 + WithTx()
│   │   ├── gorm_repo.go               # GORM 实现（ctxDB 从 context 提取事务）
│   │   └── sql_queries.go             # 复杂 SQL 常量
│   ├── service/                       # 业务逻辑层（核心引擎）
│   │   ├── engine.go                  # Engine 结构体 + NewEngine()
│   │   ├── process_define.go          # 流程定义解析/保存
│   │   ├── process_instance.go        # 流程实例管理
│   │   ├── process_node.go            # 节点调度处理
│   │   ├── task_service.go            # 任务操作
│   │   ├── event_system.go            # 事件系统（反射）
│   │   ├── variable.go                # 变量系统
│   │   ├── expression.go              # 表达式求值（expr-lang/expr）
│   │   └── scheduler.go               # 计划任务
│   ├── pkg/                           # 内部工具
│   │   ├── helper.go                  # JSON、去重等通用函数
│   │   └── reflect.go                 # 反射工具
│   └── web/                           # 可选 Web API 模块
│       ├── server.go                  # WebServer + StartWebAPI()
│       ├── router.go                  # Gin 路由注册
│       └── handler/                   # HTTP 处理器（ShouldBind + 请求 struct）
│           ├── proc_def.go
│           ├── proc_inst.go
│           └── task.go
└── example/                           # 使用示例
    ├── main.go
    ├── event/example_event.go
    ├── process/example_process.go
    └── schedule/example_schedule.go
```

### 层间依赖关系

```
engine.go (根包公共API)
  └── internal/service (业务逻辑)
        ├── internal/repository (数据访问)
        │     └── internal/entity (GORM 模型)
        └── internal/model (领域模型)
  └── internal/web (可选Web API)
        └── internal/service
```

依赖方向：`根包 → service → repository → entity`，`根包 → web → service`，所有层可依赖 `model` 和 `pkg`。

### 使用方式

```go
import (
    "context"

    "github.com/Bunny3th/easy-workflow"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

// 调用方负责创建和配置 GORM 连接（连接池、日志等）
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{...})
if err != nil {
    log.Fatal(err)
}

// NewEngine 接收外部 db，负责 AutoMigrate 和内部组件初始化
eng, err := easyworkflow.New(db, easyworkflow.Config{
    IgnoreEventError: false,
})
if err != nil {
    log.Fatal(err)
}

eng.RegisterEvents(&MyEvent{})
ctx := context.Background()
id, err := eng.InstanceStart(ctx, procID, businessID, comment, variablesJSON)
err = eng.TaskPass(ctx, taskID, comment, variablesJSON, false)

// 可选：启动 Web API
eng.StartWebAPI(ginEngine, easyworkflow.WebConfig{
    BaseURL: "/process",
    Addr:    ":8080",
})
```

### context.Context 传递

所有 service/repository 方法第一个参数为 `ctx context.Context`。Web handler 通过 `c.Request.Context()` 传递。事件回调内部使用 `context.Background()`（事件签名保持不变以兼容反射）。

### 事务管理

Service 层使用 GORM 的 `db.Transaction()` 管理事务，不手动调用 `Begin/Commit/Rollback`。事务通过 `context` 传递给 Repository：

```go
// service 层
err := e.db.Transaction(func(tx *gorm.DB) error {
    txCtx := repository.WithTx(ctx, tx)
    if err := e.repo.CreateInstance(txCtx, &inst); err != nil {
        return err  // 返回 error 自动回滚
    }
    return nil  // 返回 nil 自动提交
})

// repository 层通过 ctxDB 从 context 提取连接
func (r *GormRepo) ctxDB(ctx context.Context) *gorm.DB {
    if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
        return tx  // 事务内
    }
    return r.db  // 默认连接
}
```

Repository 接口的所有方法均接受 `ctx context.Context`，**不**接受 `*gorm.DB` 参数。需要事务时，通过 `repository.WithTx(ctx, tx)` 将事务放入 context。

### 请求参数结构化

Web handler 使用 `c.ShouldBind(&req)` 绑定请求参数，参数定义在 `internal/model/request.go`：

| 结构体 | 用途 | 嵌套 |
|--------|------|------|
| `PageQuery` | 分页（pageNo/pageSize，带默认值） | 被 TaskListReq、InstanceListReq 嵌套 |
| `PageData[T]` | 泛型分页响应 | |
| `TaskActionReq` | Pass/Reject 共享 | 被 TaskFreeRejectReq 嵌套 |
| `TaskFreeRejectReq` | 自由驳回 | 嵌套 TaskActionReq |
| `TaskTransferReq` | 任务转交 | |
| `TaskInfoReq` | 任务信息查询 | |
| `TaskListReq` | 待办任务列表 | 嵌套 PageQuery |
| `TaskFinishedListReq` | 已办任务列表 | 嵌套 TaskListReq |
| `InstanceStartReq` | 启动流程实例 | |
| `InstanceRevokeReq` | 撤销流程实例 | |
| `InstanceListReq` | 流程实例列表 | 嵌套 PageQuery |
| `ProcessSaveReq` | 保存流程定义 | |
| `ProcessListReq` | 流程定义列表 | |
| `ProcessDefGetReq` | 获取流程定义 | |

### 核心概念与数据流

**流程定义 (Process)** → JSON 格式存储在 `proc_def.resource` 字段中，包含流程名、来源、撤销事件和节点数组。

**四种节点类型** (`model.NodeType`):
- `0` (RootNode): 开始节点 — 特殊任务节点，启动后自动通过
- `1` (TaskNode): 任务节点 — 需要人处理的审批节点
- `2` (GateWayNode): 网关节点 — 混合网关（排他+并行+包含）
- `3` (EndNode): 结束节点 — 触发数据归档

**流程执行流程**: `processNode()` (`internal/service/process_node.go`) 是核心调度函数，根据节点类型分发处理：
- RootNode → `taskNodeHandle()` 生成任务
- GateWayNode → `gatewayNodeHandle()` 条件判断+分支
- TaskNode → `taskNodeHandle()` 生成任务
- EndNode → `endNodeHandle()` 数据归档

### 引擎初始化 (`easyworkflow.New`)

`easyworkflow.New(db *gorm.DB, cfg Config)` 接受由调用方创建的 `*gorm.DB`，引擎内部负责：
1. AutoMigrate 创建/更新所有表
2. 创建 Repository、表达式求值器
3. 初始化事件池、流程缓存、计划任务池

数据库连接的创建、连接池配置、GORM 日志等由调用方在传入 `*gorm.DB` 前自行设置。`Config` 仅包含业务配置（`IgnoreEventError`、`Logger`）。`Close()` 不关闭数据库连接，由调用方负责。

### 任务处理 (`internal/service/task_service.go`)

`processTask()` 是任务审批的核心函数，处理通过/驳回：
- `TaskPass()` / `TaskReject()` → `processTask()` → `taskSubmit()` → 事件处理 → 获取下一节点 → `processNode()`
- 如果后处理失败，调用 `taskRevoke()` 回滚任务状态

**关键特性**:
- **会签** (`IsCosigned=1`): 全部通过才通过，一人驳回即驳回
- **自由驳回**: `TaskFreeReject()` 可驳回到任意上游节点
- **DirectlyToWhoRejectedMe**: 通过后直接跳到上一个驳回自己的节点（非会签节点可用）
- **任务转交**: `TaskTransfer()` 删除原任务，为新用户生成新任务

### 事件系统 (`internal/service/event_system.go`)

使用反射实现事件注册和调用：
- 事件方法定义在用户自定义 struct 上（必须是指针接收者）
- `RegisterEvents()` 传入 struct 指针，通过反射注册到 `eventPool`
- 如果 struct 具有 `SetEngine(*service.Engine)` 方法，注册时自动注入引擎引用
- **节点事件签名**: `func(e *Struct, ProcessInstanceID int, CurrentNode *model.Node, PrevNode model.Node) error`
- **流程撤销事件签名**: `func(e *Struct, ProcessInstanceID int, RevokeUserID string) error`
- 四种事件：NodeStartEvents、NodeEndEvents、TaskFinishEvents、RevokeEvents

### 混合网关 (`model.HybridGateway`, `internal/service/process_node.go`)

`HybridGateway` 是三种网关的统一实现：
- `WaitForAllPrevNode=1`: 并行模式，所有上级节点完成才继续
- `WaitForAllPrevNode=0`: 包含模式，任一上级节点完成即可
- `Conditions`: 条件表达式数组，用 `$变量名` 引用流程变量，通过 `expr-lang/expr` 计算表达式结果
- `InevitableNodes`: 必然执行的节点列表

### 表达式求值 (`internal/service/expression.go`)

使用 `github.com/expr-lang/expr` 库实现内存安全求值：
- 变量保持 `$prefix` 语法（如 `$days>=3`），内部自动剥离 `$` 前缀
- 自动类型转换：string → int/float/bool（支持数值比较）
- 支持 `>=`、`<=`、`==`、`!=`、`>`、`<`、`and/&&`、`or/||` 等运算符
- 编译安全，无 SQL 注入风险

### 变量系统 (`internal/service/variable.go`)

- 流程变量以 `$key` 格式引用，存储在 `proc_inst_variable` 表
- `ResolveVariables()` 解析变量为实际值
- 节点 UserIDs 中可用变量引用（如 `$starter`）

### 计划任务 (`internal/service/scheduler.go`)

- `ScheduleTask()` 注册定时任务，内部管理调度
- 支持设置开始/结束时间和执行间隔

### 数据库表

所有业务表都有对应的历史表（`hist_` 前缀），流程结束时数据归档到历史表。核心表：
- `proc_def` / `hist_proc_def` — 流程定义
- `proc_inst` / `hist_proc_inst` — 流程实例
- `proc_task` / `hist_proc_task` — 任务
- `proc_execution` / `hist_proc_execution` — 节点执行关系
- `proc_inst_variable` / `hist_proc_inst_variable` — 流程变量

### Web API (`internal/web/`)

基于 Gin，所有 API 在 `internal/web/router.go` 中定义。API 前缀可配置。端点分三组：
- `/def/*` — 流程定义管理
- `/inst/*` — 流程实例操作
- `/task/*` — 任务操作

Handler 通过 struct 持有 `*service.Engine` 引用，使用 `c.ShouldBind(&req)` 绑定请求参数到 `model/request.go` 中定义的结构体。

### 流程定义缓存

`procCache` (`internal/service/engine.go`): `map[int]map[string]model.Node`，缓存流程节点定义避免重复查询。流程保存时自动清除对应缓存。

## Key Conventions

- **禁止 dot import**：所有包使用显式包名引用
- **struct-based 设计**：所有状态封装在结构体中，无全局可变变量
- **手动依赖注入**：通过构造函数注入依赖，不使用 Wire。`NewEngine(db, cfg)` 接受外部 `*gorm.DB`
- **context.Context**：所有 service/repository 方法第一个参数为 `ctx context.Context`
- **事务管理**：Service 层使用 `db.Transaction()`，Repository 通过 `repository.WithTx(ctx, tx)` 从 context 获取事务连接
- **请求参数结构化**：Handler 使用 `model/request.go` 中定义的 struct + `ShouldBind` 绑定参数
- **数据库访问**：通过 `Repository` 接口抽象，实现层通过 `ctxDB(ctx)` 从 context 获取连接，混合使用 GORM ORM 和原生 SQL
- **日志**：统一使用 `slog`，引擎接受可选的 `*slog.Logger`，默认使用 `slog.Default()`
- **注释**：全部使用中文注释，导出符号使用 godoc 格式
- **文件命名**：全部使用 snake_case
- **函数/参数命名**：导出符号 PascalCase，未导出符号 camelCase，参数 camelCase
- **流程定义**以 JSON 形式存储在数据库中，通过 `ProcessParse()` 反序列化
- **复杂 SQL** 提取为常量定义在 `internal/repository/sql_queries.go` 中
