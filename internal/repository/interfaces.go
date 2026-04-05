package repository

import (
	"context"

	"github.com/Bunny3th/easy-workflow/internal/entity"
	"github.com/Bunny3th/easy-workflow/internal/model"
	"gorm.io/gorm"
)

// txKey 用于从 context 中获取事务连接。
type txKey struct{}

// WithTx 将事务 db 放入 context，供 Repository 方法使用。
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// Repository 数据访问层接口，封装所有数据库操作。
// 所有方法通过 ctx 传递 context 和事务连接。
// Service 层使用 db.Transaction 时，通过 WithTx 将事务放入 ctx 传入。
type Repository interface {
	// DB 返回底层数据库连接，用于 service 层管理事务。
	DB() *gorm.DB

	// ---- 流程定义 ----

	GetProcessIDByName(ctx context.Context, p GetProcIDByNameParams) (procID int, version int, err error)
	GetProcessResource(ctx context.Context, procID int) (string, error)
	ListProcessDef(ctx context.Context, source string) ([]entity.ProcDef, error)
	ArchiveProcDef(ctx context.Context, name, source string) error
	UpdateProcDef(ctx context.Context, p UpdateProcDefParams) error
	ArchiveExecutions(ctx context.Context, procID int) error
	DeleteExecutions(ctx context.Context, procID int) error
	SaveExecutions(ctx context.Context, executions []entity.ProcExecution) error
	CreateProcDef(ctx context.Context, procDef *entity.ProcDef) error

	// ---- 流程实例 ----

	CreateInstance(ctx context.Context, inst *entity.ProcInst) error
	UpdateInstance(ctx context.Context, id int, updates map[string]any) error
	GetInstanceInfo(ctx context.Context, instID int) (model.InstanceView, error)
	ListInstanceStartByUser(ctx context.Context, p ListInstByUserParams) ([]model.InstanceView, error)
	CountInstanceStartByUser(ctx context.Context, p CountByUserParams) (int64, error)
	GetProcessIDByInstID(ctx context.Context, instID int) (int, error)
	GetProcessNameByInstID(ctx context.Context, instID int) (string, error)

	// ---- 任务 ----

	CreateTasks(ctx context.Context, tasks []entity.ProcTask) error
	UpdateTask(ctx context.Context, id int, updates map[string]any) error
	GetTaskInfo(ctx context.Context, taskID int) (model.TaskView, error)
	ListTaskToDo(ctx context.Context, p ListToDoParams) ([]model.TaskView, error)
	CountTaskToDo(ctx context.Context, p CountByUserParams) (int64, error)
	ListTaskFinished(ctx context.Context, p ListFinishedParams) ([]model.TaskView, error)
	CountTaskFinished(ctx context.Context, p CountFinishedParams) (int64, error)
	ListInstanceTaskHistory(ctx context.Context, instID int) ([]model.TaskView, error)
	GetTaskNodeStatus(ctx context.Context, p TaskNodeStatusParams) (total, passed, rejected int, err error)
	GetNotFinishUsers(ctx context.Context, p NotFinishUsersParams) ([]string, error)
	GetPrevNodeBatchCode(ctx context.Context, taskID int) (string, error)
	HasRejectInBatch(ctx context.Context, batchCode string) (bool, error)
	UpdateTasksByBatchCode(ctx context.Context, batchCode string, updates map[string]any) error
	DeleteTasksByBatchCode(ctx context.Context, batchCode string) error
	DeleteTaskByID(ctx context.Context, taskID int) error
	RevokeTask(ctx context.Context, taskID int) error
	GetNextNodeIDByPrevNodeID(ctx context.Context, prevNodeID string) (string, error)
	GetUpstreamNodes(ctx context.Context, nodeID string) ([]model.Node, error)

	// ---- 执行关系 ----

	GetStartNodeID(ctx context.Context, procID int) (string, error)
	IsNodeFinished(ctx context.Context, p IsNodeFinishedParams) (bool, error)

	// ---- 变量 ----

	// SaveVariable 保存变量，如果变量不存在则创建，存在则更新。
	SaveVariable(ctx context.Context, instID int, variables []model.Variable) error
	// GetVariable 获取变量值，如果变量不存在则返回空字符串
	GetVariable(ctx context.Context, instID int, key string) (string, bool, error)
	// ResolveVariables 解析变量，获取并设置其 value，返回 map（非变量则原样存储）。
	ResolveVariables(ctx context.Context, instID int, varNames []string) (map[string]string, error)

	// ---- 归档 ----

	ArchiveInstance(ctx context.Context, instID int, status int) error
}
