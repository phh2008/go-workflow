package service

import (
	"context"

	"github.com/Bunny3th/easy-workflow/internal/entity"
	"github.com/Bunny3th/easy-workflow/internal/model"
	"gorm.io/gorm"
)

// mockRepo 是 repository.Repository 接口的 mock 实现，仅用于测试。
type mockRepo struct {
	// 流程定义
	GetProcessIDByNameFunc func(ctx context.Context, name, source string) (int, int, error)
	GetProcessResourceFunc func(ctx context.Context, procID int) (string, error)
	ListProcessDefFunc     func(ctx context.Context, source string) ([]entity.ProcDef, error)
	ArchiveProcDefFunc     func(ctx context.Context, name, source string) error
	UpdateProcDefFunc      func(ctx context.Context, name, source string, resource, userID string, version int) error
	ArchiveExecutionsFunc  func(ctx context.Context, procID int) error
	DeleteExecutionsFunc   func(ctx context.Context, procID int) error
	SaveExecutionsFunc     func(ctx context.Context, executions []entity.ProcExecution) error
	CreateProcDefFunc      func(ctx context.Context, procDef *entity.ProcDef) error

	// 流程实例
	CreateInstanceFunc          func(ctx context.Context, inst *entity.ProcInst) error
	UpdateInstanceFunc          func(ctx context.Context, id int, updates map[string]any) error
	GetInstanceInfoFunc         func(ctx context.Context, instID int) (model.InstanceView, error)
	ListInstanceStartByUserFunc     func(ctx context.Context, userID, processName string, offset, limit int) ([]model.InstanceView, error)
	CountInstanceStartByUserFunc    func(ctx context.Context, userID, processName string) (int64, error)
	GetProcessIDByInstIDFunc        func(ctx context.Context, instID int) (int, error)
	GetProcessNameByInstIDFunc      func(ctx context.Context, instID int) (string, error)

	// 任务
	CreateTasksFunc               func(ctx context.Context, tasks []entity.ProcTask) error
	UpdateTaskFunc                func(ctx context.Context, id int, updates map[string]any) error
	GetTaskInfoFunc               func(ctx context.Context, taskID int) (model.TaskView, error)
	ListTaskToDoFunc              func(ctx context.Context, userID, processName string, asc bool, offset, limit int) ([]model.TaskView, error)
	CountTaskToDoFunc             func(ctx context.Context, userID, processName string) (int64, error)
	ListTaskFinishedFunc          func(ctx context.Context, userID, processName string, ignoreStartByMe, asc bool, offset, limit int) ([]model.TaskView, error)
	CountTaskFinishedFunc         func(ctx context.Context, userID, processName string, ignoreStartByMe bool) (int64, error)
	ListInstanceTaskHistoryFunc   func(ctx context.Context, instID int) ([]model.TaskView, error)
	GetTaskNodeStatusFunc         func(ctx context.Context, instID int, nodeID, batchCode string) (int, int, int, error)
	GetNotFinishUsersFunc         func(ctx context.Context, instID int, nodeID string) ([]string, error)
	GetPrevNodeBatchCodeFunc      func(ctx context.Context, taskID int) (string, error)
	HasRejectInBatchFunc          func(ctx context.Context, batchCode string) (bool, error)
	UpdateTasksByBatchCodeFunc    func(ctx context.Context, batchCode string, updates map[string]any) error
	DeleteTasksByBatchCodeFunc    func(ctx context.Context, batchCode string) error
	DeleteTaskByIDFunc            func(ctx context.Context, taskID int) error
	RevokeTaskFunc                func(ctx context.Context, taskID int) error
	GetNextNodeIDByPrevNodeIDFunc func(ctx context.Context, prevNodeID string) (string, error)
	GetUpstreamNodesFunc          func(ctx context.Context, nodeID string) ([]model.Node, error)

	// 执行关系
	GetStartNodeIDFunc func(ctx context.Context, procID int) (string, error)
	IsNodeFinishedFunc func(ctx context.Context, instID, nodeID string) (bool, error)

	// 变量
	SaveVariableFunc     func(ctx context.Context, instID int, variables []model.Variable) error
	GetVariableFunc      func(ctx context.Context, instID int, key string) (string, bool, error)
	ResolveVariablesFunc func(ctx context.Context, instID int, varNames []string) (map[string]string, error)

	// 归档
	ArchiveInstanceFunc func(ctx context.Context, instID int, status int) error
}

func (m *mockRepo) DB() *gorm.DB { return nil }
func (m *mockRepo) GetProcessIDByName(ctx context.Context, name, source string) (int, int, error) {
	if m.GetProcessIDByNameFunc != nil {
		return m.GetProcessIDByNameFunc(ctx, name, source)
	}
	return 0, 0, nil
}
func (m *mockRepo) GetProcessResource(ctx context.Context, procID int) (string, error) {
	if m.GetProcessResourceFunc != nil {
		return m.GetProcessResourceFunc(ctx, procID)
	}
	return "", nil
}
func (m *mockRepo) ListProcessDef(ctx context.Context, source string) ([]entity.ProcDef, error) {
	if m.ListProcessDefFunc != nil {
		return m.ListProcessDefFunc(ctx, source)
	}
	return nil, nil
}
func (m *mockRepo) ArchiveProcDef(ctx context.Context, name, source string) error {
	if m.ArchiveProcDefFunc != nil {
		return m.ArchiveProcDefFunc(ctx, name, source)
	}
	return nil
}
func (m *mockRepo) UpdateProcDef(ctx context.Context, name, source string, resource, userID string, version int) error {
	if m.UpdateProcDefFunc != nil {
		return m.UpdateProcDefFunc(ctx, name, source, resource, userID, version)
	}
	return nil
}
func (m *mockRepo) ArchiveExecutions(ctx context.Context, procID int) error {
	if m.ArchiveExecutionsFunc != nil {
		return m.ArchiveExecutionsFunc(ctx, procID)
	}
	return nil
}
func (m *mockRepo) DeleteExecutions(ctx context.Context, procID int) error {
	if m.DeleteExecutionsFunc != nil {
		return m.DeleteExecutionsFunc(ctx, procID)
	}
	return nil
}
func (m *mockRepo) SaveExecutions(ctx context.Context, executions []entity.ProcExecution) error {
	if m.SaveExecutionsFunc != nil {
		return m.SaveExecutionsFunc(ctx, executions)
	}
	return nil
}
func (m *mockRepo) CreateProcDef(ctx context.Context, procDef *entity.ProcDef) error {
	if m.CreateProcDefFunc != nil {
		return m.CreateProcDefFunc(ctx, procDef)
	}
	return nil
}
func (m *mockRepo) CreateInstance(ctx context.Context, inst *entity.ProcInst) error {
	if m.CreateInstanceFunc != nil {
		return m.CreateInstanceFunc(ctx, inst)
	}
	return nil
}
func (m *mockRepo) UpdateInstance(ctx context.Context, id int, updates map[string]any) error {
	if m.UpdateInstanceFunc != nil {
		return m.UpdateInstanceFunc(ctx, id, updates)
	}
	return nil
}
func (m *mockRepo) GetInstanceInfo(ctx context.Context, instID int) (model.InstanceView, error) {
	if m.GetInstanceInfoFunc != nil {
		return m.GetInstanceInfoFunc(ctx, instID)
	}
	return model.InstanceView{}, nil
}
func (m *mockRepo) ListInstanceStartByUser(ctx context.Context, userID, processName string, offset, limit int) ([]model.InstanceView, error) {
	if m.ListInstanceStartByUserFunc != nil {
		return m.ListInstanceStartByUserFunc(ctx, userID, processName, offset, limit)
	}
	return nil, nil
}
func (m *mockRepo) CountInstanceStartByUser(ctx context.Context, userID, processName string) (int64, error) {
	if m.CountInstanceStartByUserFunc != nil {
		return m.CountInstanceStartByUserFunc(ctx, userID, processName)
	}
	return 0, nil
}
func (m *mockRepo) GetProcessIDByInstID(ctx context.Context, instID int) (int, error) {
	if m.GetProcessIDByInstIDFunc != nil {
		return m.GetProcessIDByInstIDFunc(ctx, instID)
	}
	return 0, nil
}
func (m *mockRepo) GetProcessNameByInstID(ctx context.Context, instID int) (string, error) {
	if m.GetProcessNameByInstIDFunc != nil {
		return m.GetProcessNameByInstIDFunc(ctx, instID)
	}
	return "", nil
}
func (m *mockRepo) CreateTasks(ctx context.Context, tasks []entity.ProcTask) error {
	if m.CreateTasksFunc != nil {
		return m.CreateTasksFunc(ctx, tasks)
	}
	return nil
}
func (m *mockRepo) UpdateTask(ctx context.Context, id int, updates map[string]any) error {
	if m.UpdateTaskFunc != nil {
		return m.UpdateTaskFunc(ctx, id, updates)
	}
	return nil
}
func (m *mockRepo) GetTaskInfo(ctx context.Context, taskID int) (model.TaskView, error) {
	if m.GetTaskInfoFunc != nil {
		return m.GetTaskInfoFunc(ctx, taskID)
	}
	return model.TaskView{}, nil
}
func (m *mockRepo) ListTaskToDo(ctx context.Context, userID, processName string, asc bool, offset, limit int) ([]model.TaskView, error) {
	if m.ListTaskToDoFunc != nil {
		return m.ListTaskToDoFunc(ctx, userID, processName, asc, offset, limit)
	}
	return nil, nil
}
func (m *mockRepo) CountTaskToDo(ctx context.Context, userID, processName string) (int64, error) {
	if m.CountTaskToDoFunc != nil {
		return m.CountTaskToDoFunc(ctx, userID, processName)
	}
	return 0, nil
}
func (m *mockRepo) ListTaskFinished(ctx context.Context, userID, processName string, ignoreStartByMe, asc bool, offset, limit int) ([]model.TaskView, error) {
	if m.ListTaskFinishedFunc != nil {
		return m.ListTaskFinishedFunc(ctx, userID, processName, ignoreStartByMe, asc, offset, limit)
	}
	return nil, nil
}
func (m *mockRepo) CountTaskFinished(ctx context.Context, userID, processName string, ignoreStartByMe bool) (int64, error) {
	if m.CountTaskFinishedFunc != nil {
		return m.CountTaskFinishedFunc(ctx, userID, processName, ignoreStartByMe)
	}
	return 0, nil
}
func (m *mockRepo) ListInstanceTaskHistory(ctx context.Context, instID int) ([]model.TaskView, error) {
	if m.ListInstanceTaskHistoryFunc != nil {
		return m.ListInstanceTaskHistoryFunc(ctx, instID)
	}
	return nil, nil
}
func (m *mockRepo) GetTaskNodeStatus(ctx context.Context, instID int, nodeID, batchCode string) (int, int, int, error) {
	if m.GetTaskNodeStatusFunc != nil {
		return m.GetTaskNodeStatusFunc(ctx, instID, nodeID, batchCode)
	}
	return 0, 0, 0, nil
}
func (m *mockRepo) GetNotFinishUsers(ctx context.Context, instID int, nodeID string) ([]string, error) {
	if m.GetNotFinishUsersFunc != nil {
		return m.GetNotFinishUsersFunc(ctx, instID, nodeID)
	}
	return nil, nil
}
func (m *mockRepo) GetPrevNodeBatchCode(ctx context.Context, taskID int) (string, error) {
	if m.GetPrevNodeBatchCodeFunc != nil {
		return m.GetPrevNodeBatchCodeFunc(ctx, taskID)
	}
	return "", nil
}
func (m *mockRepo) HasRejectInBatch(ctx context.Context, batchCode string) (bool, error) {
	if m.HasRejectInBatchFunc != nil {
		return m.HasRejectInBatchFunc(ctx, batchCode)
	}
	return false, nil
}
func (m *mockRepo) UpdateTasksByBatchCode(ctx context.Context, batchCode string, updates map[string]any) error {
	if m.UpdateTasksByBatchCodeFunc != nil {
		return m.UpdateTasksByBatchCodeFunc(ctx, batchCode, updates)
	}
	return nil
}
func (m *mockRepo) DeleteTasksByBatchCode(ctx context.Context, batchCode string) error {
	if m.DeleteTasksByBatchCodeFunc != nil {
		return m.DeleteTasksByBatchCodeFunc(ctx, batchCode)
	}
	return nil
}
func (m *mockRepo) DeleteTaskByID(ctx context.Context, taskID int) error {
	if m.DeleteTaskByIDFunc != nil {
		return m.DeleteTaskByIDFunc(ctx, taskID)
	}
	return nil
}
func (m *mockRepo) RevokeTask(ctx context.Context, taskID int) error {
	if m.RevokeTaskFunc != nil {
		return m.RevokeTaskFunc(ctx, taskID)
	}
	return nil
}
func (m *mockRepo) GetNextNodeIDByPrevNodeID(ctx context.Context, prevNodeID string) (string, error) {
	if m.GetNextNodeIDByPrevNodeIDFunc != nil {
		return m.GetNextNodeIDByPrevNodeIDFunc(ctx, prevNodeID)
	}
	return "", nil
}
func (m *mockRepo) GetUpstreamNodes(ctx context.Context, nodeID string) ([]model.Node, error) {
	if m.GetUpstreamNodesFunc != nil {
		return m.GetUpstreamNodesFunc(ctx, nodeID)
	}
	return nil, nil
}
func (m *mockRepo) GetStartNodeID(ctx context.Context, procID int) (string, error) {
	if m.GetStartNodeIDFunc != nil {
		return m.GetStartNodeIDFunc(ctx, procID)
	}
	return "", nil
}
func (m *mockRepo) IsNodeFinished(ctx context.Context, instID, nodeID string) (bool, error) {
	if m.IsNodeFinishedFunc != nil {
		return m.IsNodeFinishedFunc(ctx, instID, nodeID)
	}
	return true, nil
}
func (m *mockRepo) SaveVariable(ctx context.Context, instID int, variables []model.Variable) error {
	if m.SaveVariableFunc != nil {
		return m.SaveVariableFunc(ctx, instID, variables)
	}
	return nil
}
func (m *mockRepo) GetVariable(ctx context.Context, instID int, key string) (string, bool, error) {
	if m.GetVariableFunc != nil {
		return m.GetVariableFunc(ctx, instID, key)
	}
	return "", false, nil
}
func (m *mockRepo) ResolveVariables(ctx context.Context, instID int, varNames []string) (map[string]string, error) {
	if m.ResolveVariablesFunc != nil {
		return m.ResolveVariablesFunc(ctx, instID, varNames)
	}
	return make(map[string]string), nil
}
func (m *mockRepo) ArchiveInstance(ctx context.Context, instID int, status int) error {
	if m.ArchiveInstanceFunc != nil {
		return m.ArchiveInstanceFunc(ctx, instID, status)
	}
	return nil
}
