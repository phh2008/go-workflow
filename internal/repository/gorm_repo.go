package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Bunny3th/easy-workflow/internal/entity"
	"github.com/Bunny3th/easy-workflow/internal/model"
	"gorm.io/gorm"
)

// GormRepo 基于 GORM 的 Repository 实现。
type GormRepo struct {
	db *gorm.DB
}

// NewGormRepo 创建 GormRepo 实例。
func NewGormRepo(db *gorm.DB) *GormRepo {
	return &GormRepo{db: db}
}

// DB 返回底层数据库连接。
func (r *GormRepo) DB() *gorm.DB {
	return r.db
}

// ctxDB 从 context 中提取事务连接，如果不存在则返回默认连接。
func (r *GormRepo) ctxDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return r.db
}

// ===================== 流程定义 =====================

// GetProcessIDByName 根据流程名称和来源获取流程ID和版本号。
func (r *GormRepo) GetProcessIDByName(ctx context.Context, name, source string) (int, int, error) {
	type result struct {
		ID      int
		Version int
	}
	var res result
	if err := r.ctxDB(ctx).Raw(sqlGetProcessIDByName, name, source).Scan(&res).Error; err != nil {
		return 0, 0, err
	}
	return res.ID, res.Version, nil
}

// GetProcessResource 根据流程ID获取流程定义资源（JSON）。
func (r *GormRepo) GetProcessResource(ctx context.Context, procID int) (string, error) {
	type res struct {
		Resource string
	}
	var r2 res
	if err := r.ctxDB(ctx).Raw(sqlGetProcessResource, procID).Scan(&r2).Error; err != nil {
		return "", err
	}
	if r2.Resource == "" {
		return "", errors.New("未找到对应流程定义")
	}
	return r2.Resource, nil
}

// ListProcessDef 获取某个 source 下所有流程定义。
func (r *GormRepo) ListProcessDef(ctx context.Context, source string) ([]entity.ProcDef, error) {
	var procDefs []entity.ProcDef
	if err := r.ctxDB(ctx).Raw(sqlListProcessDef, source).Scan(&procDefs).Error; err != nil {
		return nil, err
	}
	return procDefs, nil
}

// ArchiveProcDef 将老版本流程定义移到历史表中。
func (r *GormRepo) ArchiveProcDef(ctx context.Context, name, source string) error {
	return r.ctxDB(ctx).Exec(sqlArchiveProcDef, name, source).Error
}

// UpdateProcDef 更新现有流程定义。
func (r *GormRepo) UpdateProcDef(ctx context.Context, name, source string, resource, userID string, version int) error {
	return r.ctxDB(ctx).Model(&entity.ProcDef{}).
		Where("name=? AND source=?", name, source).
		Updates(entity.ProcDef{
			Version:   version,
			Resource:  resource,
			UserID:    userID,
			CreatTime: entity.Now(),
		}).Error
}

// ArchiveExecutions 将指定流程ID的执行关系移到历史表中。
func (r *GormRepo) ArchiveExecutions(ctx context.Context, procID int) error {
	return r.ctxDB(ctx).Exec(sqlArchiveExecutions, procID).Error
}

// DeleteExecutions 删除指定流程ID的执行关系。
func (r *GormRepo) DeleteExecutions(ctx context.Context, procID int) error {
	return r.ctxDB(ctx).Where("proc_id=?", procID).Delete(&entity.ProcExecution{}).Error
}

// SaveExecutions 批量保存流程节点执行关系定义。
func (r *GormRepo) SaveExecutions(ctx context.Context, executions []entity.ProcExecution) error {
	return r.ctxDB(ctx).Create(&executions).Error
}

// CreateProcDef 创建新的流程定义记录。
func (r *GormRepo) CreateProcDef(ctx context.Context, procDef *entity.ProcDef) error {
	return r.ctxDB(ctx).Create(&procDef).Error
}

// ===================== 流程实例 =====================

// CreateInstance 创建流程实例记录。
func (r *GormRepo) CreateInstance(ctx context.Context, inst *entity.ProcInst) error {
	return r.ctxDB(ctx).Create(&inst).Error
}

// UpdateInstance 更新流程实例指定字段。
func (r *GormRepo) UpdateInstance(ctx context.Context, id int, updates map[string]any) error {
	return r.ctxDB(ctx).Model(&entity.ProcInst{}).Where("id=?", id).Updates(updates).Error
}

// GetInstanceInfo 通过 CTE 合并当前表和历史表获取流程实例信息。
func (r *GormRepo) GetInstanceInfo(ctx context.Context, instID int) (model.InstanceView, error) {
	var inst model.InstanceView
	if err := r.ctxDB(ctx).Raw(sqlGetInstanceInfo, instID, instID).Scan(&inst).Error; err != nil {
		return inst, err
	}
	return inst, nil
}

// ListInstanceStartByUser 获取特定用户发起的流程实例列表（含分页）。
func (r *GormRepo) ListInstanceStartByUser(ctx context.Context, userID, processName string, offset, limit int) ([]model.InstanceView, error) {
	var instances []model.InstanceView
	condition := map[string]any{
		"userid":   userID,
		"procname": processName,
		"index":    offset,
		"rows":     limit,
	}
	if err := r.ctxDB(ctx).Raw(sqlListInstanceStartByUser, condition).Scan(&instances).Error; err != nil {
		return nil, err
	}
	return instances, nil
}

// GetProcessIDByInstID 根据流程实例ID获取流程ID。
func (r *GormRepo) GetProcessIDByInstID(ctx context.Context, instID int) (int, error) {
	var id int
	if err := r.ctxDB(ctx).Raw(sqlGetProcessIDByInstID, instID).Scan(&id).Error; err != nil {
		return 0, err
	}
	return id, nil
}

// GetProcessNameByInstID 根据流程实例ID获取流程名称。
func (r *GormRepo) GetProcessNameByInstID(ctx context.Context, instID int) (string, error) {
	type res struct {
		Name string
	}
	var r2 res
	if err := r.ctxDB(ctx).Raw(sqlGetProcessNameByInstID, instID).Scan(&r2).Error; err != nil {
		return "", err
	}
	return r2.Name, nil
}

// ===================== 任务 =====================

// CreateTasks 批量创建任务记录。
func (r *GormRepo) CreateTasks(ctx context.Context, tasks []entity.ProcTask) error {
	return r.ctxDB(ctx).Create(&tasks).Error
}

// UpdateTask 更新任务指定字段。
func (r *GormRepo) UpdateTask(ctx context.Context, id int, updates map[string]any) error {
	return r.ctxDB(ctx).Model(&entity.ProcTask{}).Where("id=?", id).Updates(updates).Error
}

// GetTaskInfo 通过 CTE 合并当前表和历史表获取任务信息。
func (r *GormRepo) GetTaskInfo(ctx context.Context, taskID int) (model.TaskView, error) {
	var task model.TaskView
	if err := r.ctxDB(ctx).Raw(sqlGetTaskInfo, taskID, taskID).Scan(&task).Error; err != nil {
		return model.TaskView{}, err
	}
	if task.TaskID == 0 {
		return model.TaskView{}, fmt.Errorf("ID为%d的任务不存在!", taskID)
	}
	return task, nil
}

// ListTaskToDo 获取特定用户待办任务列表。
func (r *GormRepo) ListTaskToDo(ctx context.Context, userID, processName string, asc bool, offset, limit int) ([]model.TaskView, error) {
	var tasks []model.TaskView
	sortBy := "DESC"
	if asc {
		sortBy = "ASC"
	}
	sql := strings.ReplaceAll(sqlGetTaskToDoList, "{{SORT}}", sortBy)
	condition := map[string]any{
		"userid":   userID,
		"procname": processName,
		"index":    offset,
		"rows":     limit,
	}
	if err := r.ctxDB(ctx).Raw(sql, condition).Scan(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// ListTaskFinished 获取特定用户已完成任务列表。
func (r *GormRepo) ListTaskFinished(ctx context.Context, userID, processName string, ignoreStartByMe, asc bool, offset, limit int) ([]model.TaskView, error) {
	var tasks []model.TaskView
	sortBy := "DESC"
	if asc {
		sortBy = "ASC"
	}
	// 当传入 UserID 为空时，IgnoreStartByMe 参数强制变为 False
	if userID == "" {
		ignoreStartByMe = false
	}
	sql := strings.ReplaceAll(sqlGetTaskFinishedList, "{{SORT}}", sortBy)
	condition := map[string]any{
		"userid":          userID,
		"procname":        processName,
		"index":           offset,
		"rows":            limit,
		"ignorestartbyme": ignoreStartByMe,
	}
	if err := r.ctxDB(ctx).Raw(sql, condition).Scan(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// ListInstanceTaskHistory 获取流程实例下所有任务历史记录。
func (r *GormRepo) ListInstanceTaskHistory(ctx context.Context, instID int) ([]model.TaskView, error) {
	var tasks []model.TaskView
	if err := r.ctxDB(ctx).Raw(sqlGetInstanceTaskHistory, instID, instID).Scan(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTaskNodeStatus 获取任务节点审批状态，返回总任务数、通过数、驳回数。
func (r *GormRepo) GetTaskNodeStatus(ctx context.Context, instID int, nodeID, batchCode string) (int, int, int, error) {
	type res struct {
		TotalTask     int `gorm:"column:total_task"`
		TotalPassed   int `gorm:"column:total_passed"`
		TotalRejected int `gorm:"column:total_rejected"`
	}
	var result res
	if err := r.ctxDB(ctx).Raw(sqlTaskNodeStatus, instID, nodeID, batchCode).Scan(&result).Error; err != nil {
		return 0, 0, 0, err
	}
	return result.TotalTask, result.TotalPassed, result.TotalRejected, nil
}

// GetNotFinishUsers 获取某节点中未完成任务的用户ID列表。
func (r *GormRepo) GetNotFinishUsers(ctx context.Context, instID int, nodeID string) ([]string, error) {
	type notFinishUser struct {
		UserID string `gorm:"column:user_id"`
	}
	var users []notFinishUser
	if err := r.ctxDB(ctx).Raw(sqlGetNotFinishUsers, instID, nodeID).Scan(&users).Error; err != nil {
		return nil, err
	}
	result := make([]string, 0, len(users))
	for _, u := range users {
		result = append(result, u.UserID)
	}
	return result, nil
}

// GetPrevNodeBatchCode 获取任务的上一个实际执行节点的批次码。
func (r *GormRepo) GetPrevNodeBatchCode(ctx context.Context, taskID int) (string, error) {
	type batchCode struct {
		BatchCode string `gorm:"column:batch_code"`
	}
	var bc batchCode
	if err := r.ctxDB(ctx).Raw(sqlGetPrevNodeBatchCode, taskID).Scan(&bc).Error; err != nil {
		return "", err
	}
	return bc.BatchCode, nil
}

// HasRejectInBatch 检查指定批次码中是否存在驳回状态的任务。
func (r *GormRepo) HasRejectInBatch(ctx context.Context, batchCode string) (bool, error) {
	var id int
	if err := r.ctxDB(ctx).Raw(sqlHasRejectInBatch, batchCode).Scan(&id).Error; err != nil {
		return false, err
	}
	return id != 0, nil
}

// UpdateTasksByBatchCode 按批次码批量更新任务字段。
func (r *GormRepo) UpdateTasksByBatchCode(ctx context.Context, batchCode string, updates map[string]any) error {
	return r.ctxDB(ctx).Model(&entity.ProcTask{}).Where("batch_code=?", batchCode).Updates(updates).Error
}

// DeleteTasksByBatchCode 按批次码删除任务。
func (r *GormRepo) DeleteTasksByBatchCode(ctx context.Context, batchCode string) error {
	return r.ctxDB(ctx).Where("batch_code=?", batchCode).Delete(&entity.ProcTask{}).Error
}

// DeleteTaskByID 按任务ID删除任务。
func (r *GormRepo) DeleteTaskByID(ctx context.Context, taskID int) error {
	return r.ctxDB(ctx).Exec("DELETE FROM proc_task WHERE id=?", taskID).Error
}

// RevokeTask 将任务状态重置为初始（回滚已提交的任务）。
func (r *GormRepo) RevokeTask(ctx context.Context, taskID int) error {
	// 首先将 task 状态做初始化
	if err := r.ctxDB(ctx).Exec(sqlTaskRevoke, taskID).Error; err != nil {
		return err
	}
	// 对那些没有做通过或驳回、被自动设置为 finish 的 task 做初始化
	return r.ctxDB(ctx).Exec(sqlTaskRevokeBatch, taskID).Error
}

// GetNextNodeIDByPrevNodeID 获取流程定义中某节点的下一个节点ID。
func (r *GormRepo) GetNextNodeIDByPrevNodeID(ctx context.Context, prevNodeID string) (string, error) {
	var execution entity.ProcExecution
	if err := r.ctxDB(ctx).Where("prev_node_id=?", prevNodeID).First(&execution).Error; err != nil {
		return "", err
	}
	return execution.NodeID, nil
}

// GetUpstreamNodes 递归获取任务所在节点的所有上游任务节点。
func (r *GormRepo) GetUpstreamNodes(ctx context.Context, nodeID string) ([]model.Node, error) {
	var nodes []model.Node
	if err := r.ctxDB(ctx).Raw(sqlGetUpstreamNodes, nodeID, nodeID).Scan(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

// ===================== 执行关系 =====================

// GetStartNodeID 获取流程的开始节点ID。
func (r *GormRepo) GetStartNodeID(ctx context.Context, procID int) (string, error) {
	type res struct {
		NodeID string `gorm:"column:node_id"`
	}
	var r2 res
	if err := r.ctxDB(ctx).Raw(sqlGetStartNodeID, procID).Scan(&r2).Error; err != nil {
		return "", err
	}
	return r2.NodeID, nil
}

// IsNodeFinished 判断特定实例中某一个节点是否已经全部完成。
func (r *GormRepo) IsNodeFinished(ctx context.Context, instID, nodeID string) (bool, error) {
	var finished bool
	if err := r.ctxDB(ctx).Raw(sqlIsNodeFinished, instID, nodeID).Scan(&finished).Error; err != nil {
		return false, err
	}
	return finished, nil
}

// ===================== 变量 =====================

// SaveVariable 批量保存流程实例变量。已存在则更新，不存在则插入。
func (r *GormRepo) SaveVariable(ctx context.Context, instID int, variables []model.Variable) error {
	if len(variables) == 0 {
		return nil
	}
	db := r.ctxDB(ctx)
	for _, v := range variables {
		var existing entity.ProcInstVariable
		if err := db.Raw(sqlGetVariableForUpsert, instID, v.Key).Scan(&existing).Error; err != nil {
			return err
		}
		if existing.ID == 0 {
			if err := db.Create(&entity.ProcInstVariable{ProcInstID: instID, Key: v.Key, Value: v.Value}).Error; err != nil {
				return err
			}
		} else {
			if err := db.Model(&entity.ProcInstVariable{}).
				Where("proc_inst_id=? and `key`=?", instID, v.Key).Update("value", v.Value).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// GetVariable 获取流程实例变量值。返回变量值、是否存在、错误。
func (r *GormRepo) GetVariable(ctx context.Context, instID int, key string) (string, bool, error) {
	type res struct {
		Value string
	}
	var r2 res
	if err := r.ctxDB(ctx).Raw(sqlGetVariable, instID, key).Scan(&r2).Error; err != nil {
		return "", false, err
	}
	if r2.Value != "" {
		return r2.Value, true, nil
	}
	return r2.Value, false, nil
}

// ResolveVariables 批量解析变量，返回变量名到值的映射。
// 非变量（不以$开头）的值原样存储在映射中。
func (r *GormRepo) ResolveVariables(ctx context.Context, instID int, varNames []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, v := range varNames {
		if after, ok := strings.CutPrefix(v, "$"); ok {
			key := after
			value, ok, err := r.GetVariable(ctx, instID, key)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, errors.New("无法匹配变量:" + v)
			}
			result[v] = value
		} else {
			result[v] = v
		}
	}
	return result, nil
}

// ===================== 归档 =====================

// ArchiveInstance 归档流程实例（任务、实例、变量）。
// 调用方需要管理事务。
func (r *GormRepo) ArchiveInstance(ctx context.Context, instID int, status int) error {
	db := r.ctxDB(ctx)

	// 将 task 表中所有该流程未 finish 的设置为 finish
	if err := db.Exec(sqlArchiveFinishTasks, instID).Error; err != nil {
		return err
	}

	// 将 task 表中任务归档
	if err := db.Exec(sqlArchiveTasks, instID).Error; err != nil {
		return err
	}

	// 删除 task 表中历史数据
	if err := db.Exec(sqlDeleteTasks, instID).Error; err != nil {
		return err
	}

	// 更新 proc_inst 表中状态
	if err := db.Exec(sqlUpdateInstanceStatus, status, instID).Error; err != nil {
		return err
	}

	// 将 proc_inst 表中数据归档
	if err := db.Exec(sqlArchiveInstance, instID).Error; err != nil {
		return err
	}

	// 删除 proc_inst 表中历史数据
	if err := db.Exec(sqlDeleteInstance, instID).Error; err != nil {
		return err
	}

	// 将 proc_inst_variable 表中数据归档
	if err := db.Exec(sqlArchiveVariables, instID).Error; err != nil {
		return err
	}

	// 删除 proc_inst_variable 表中历史数据
	if err := db.Exec(sqlDeleteVariables, instID).Error; err != nil {
		return err
	}

	return nil
}
