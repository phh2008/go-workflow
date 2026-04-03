# 分页查询优化方案：支持总记录数

## 问题分析

当前 3 个分页接口（StartByUser、ToDoList、FinishedList）只返回当前页数据，不返回总记录数：
- `PageData[T]` 虽已定义（含 `Count` 字段），但未被任何接口使用
- Repository 只查数据，没有 COUNT 查询
- Handler 直接返回数组，未用 `PageData` 包装

## 优化方案

### 核心思路

利用 MySQL 8.0 CTE + `SQL_CALC_FOUND_ROWS` / `FOUND_ROWS()` 的替代方案——**对 CTE 做 COUNT 子查询**，在一次 SQL 中同时获取数据和总数。

具体做法：将现有的分页 SQL 包在一个外层查询中，用子查询 `SELECT COUNT(*) OVER()` 作为窗口函数附加总数字段，同时去掉 `LIMIT` 子句用于 COUNT 查询。

实际上更简洁的方式是：**写独立的 COUNT SQL**，复用与列表查询相同的 WHERE 条件。为每个分页列表新增一个 `Count*` SQL 和 Repository 方法。

### 修改清单

#### 1. `internal/repository/sql_queries.go` — 新增 3 个 COUNT SQL

- `sqlCountInstanceStartByUser` — 复用 `sqlListInstanceStartByUser` 的 CTE + WHERE，只取 `COUNT(*)`
- `sqlCountTaskToDo` — 复用 `sqlGetTaskToDoList` 的 WHERE
- `sqlCountTaskFinished` — 复用 `sqlGetTaskFinishedList` 的 CTE + WHERE

#### 2. `internal/repository/interfaces.go` — 新增 3 个 Count 方法

```go
CountInstanceStartByUser(ctx context.Context, userID, processName string) (int64, error)
CountTaskToDo(ctx context.Context, userID, processName string) (int64, error)
CountTaskFinished(ctx context.Context, userID, processName string, ignoreStartByMe bool) (int64, error)
```

#### 3. `internal/repository/gorm_repo.go` — 实现这 3 个 Count 方法

调用对应的 COUNT SQL，返回 `int64`。

#### 4. `internal/service/process_instance.go` — 修改 `GetInstanceStartByUser`

返回值改为 `(*model.PageData[model.InstanceView], error)`，内部同时调数据和 COUNT。

#### 5. `internal/service/task_service.go` — 修改 `GetTaskToDoList` / `GetTaskFinishedList`

同上模式，返回值改为 `(*model.PageData[model.TaskView], error)`。

#### 6. `engine.go`（根包）— 更新公共 API 签名

三个方法返回值从 `([]model.XXXView, error)` 改为 `(*model.PageData[model.XXXView], error)`。

#### 7. `internal/web/handler/` — 更新 Handler

StartByUser、ToDoList、FinishedList 三个 handler 直接返回 `PageData` 对象。

#### 8. `internal/service/mock_repo.go` — 新增 mock 方法

新增 3 个 Count 方法的 mock 函数字段。

#### 9. 单元测试更新

更新依赖旧签名的测试用例。

### 返回值对比

**优化前**：
```go
func (e *Engine) GetTaskToDoList(...) ([]model.TaskView, error)
// 返回: [{"id":1,...}, {"id":2,...}]
```

**优化后**：
```go
func (e *Engine) GetTaskToDoList(...) (*model.PageData[model.TaskView], error)
// 返回: {"count":128, "pageNo":1, "pageSize":10, "data":[{"id":1,...},...]}
```

### 注意事项

- `PageData` 已有 `Count`、`PageNo`、`PageSize`、`Data` 字段，无需修改
- `NewPageData` 和 `SetData` 已提供，直接使用
- Handler 层使用 `req.GetPageNo()` 和 `req.GetPageSize()` 构造 `PageData`
- Service 层将 `offset, limit` 参数改为 `pageNo, pageSize`，内部计算 offset
- 根包公共 API 参数也同步调整为 `pageNo, pageSize`
