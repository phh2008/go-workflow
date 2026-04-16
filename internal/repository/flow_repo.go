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

// FlowRepo 基于 GORM 的 Repository 实现。
type FlowRepo struct {
	db *gorm.DB
}

// NewFlowRepo 创建 FlowRepo 实例。
func NewFlowRepo(db *gorm.DB) *FlowRepo {
	return &FlowRepo{db: db}
}

// DB 返回底层数据库连接。
func (r *FlowRepo) DB() *gorm.DB {
	return r.db
}

// ctxDB 从 context 中提取事务连接，如果不存在则返回默认连接。
func (r *FlowRepo) ctxDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return r.db
}

// ===================== 流程定义 =====================

// GetProcessIDByName 根据流程名称和来源获取流程ID和版本号。
func (r *FlowRepo) GetProcessIDByName(ctx context.Context, p GetProcIDByNameParams) (int, int, error) {
	type result struct {
		ID      int
		Version int
	}
	var res result
	if err := r.ctxDB(ctx).
		Table("proc_def").
		Select("id, version").
		Where("name = ? AND source = ?", p.Name, p.Source).
		Scan(&res).Error; err != nil {
		return 0, 0, err
	}
	return res.ID, res.Version, nil
}

// GetProcessResource 根据流程ID获取流程定义资源（JSON）。
func (r *FlowRepo) GetProcessResource(ctx context.Context, procID int) (string, error) {
	type res struct {
		Resource string
	}
	var r2 res
	if err := r.ctxDB(ctx).
		Table("proc_def").
		Select("resource").
		Where("id = ?", procID).
		Scan(&r2).Error; err != nil {
		return "", err
	}
	if r2.Resource == "" {
		return "", errors.New("未找到对应流程定义")
	}
	return r2.Resource, nil
}

// ListProcessDef 获取某个 source 下所有流程定义。
func (r *FlowRepo) ListProcessDef(ctx context.Context, source string) ([]entity.ProcDef, error) {
	var procDefs []entity.ProcDef
	if err := r.ctxDB(ctx).
		Where("source = ?", source).
		Find(&procDefs).Error; err != nil {
		return nil, err
	}
	return procDefs, nil
}

// ArchiveProcDef 将老版本流程定义移到历史表中。
func (r *FlowRepo) ArchiveProcDef(ctx context.Context, name, source string) error {
	return r.ctxDB(ctx).Exec(sqlArchiveProcDef, name, source).Error
}

// UpdateProcDef 更新现有流程定义。
func (r *FlowRepo) UpdateProcDef(ctx context.Context, p UpdateProcDefParams) error {
	return r.ctxDB(ctx).Model(&entity.ProcDef{}).
		Where("name=? AND source=?", p.Name, p.Source).
		Updates(entity.ProcDef{
			BaseModel: entity.BaseModel{
				CreatedBy: p.CreatedBy,
				UpdateAt:  entity.Now(),
			},
			Version:  p.Version,
			Resource: p.Resource,
		}).Error
}

// ArchiveExecutions 将指定流程ID的执行关系移到历史表中。
func (r *FlowRepo) ArchiveExecutions(ctx context.Context, procID int) error {
	return r.ctxDB(ctx).Exec(sqlArchiveExecutions, procID).Error
}

// DeleteExecutions 删除指定流程ID的执行关系。
func (r *FlowRepo) DeleteExecutions(ctx context.Context, procID int) error {
	return r.ctxDB(ctx).Where("proc_id=?", procID).Delete(&entity.ProcExecution{}).Error
}

// SaveExecutions 批量保存流程节点执行关系定义。
func (r *FlowRepo) SaveExecutions(ctx context.Context, executions []entity.ProcExecution) error {
	return r.ctxDB(ctx).Create(&executions).Error
}

// CreateProcDef 创建新的流程定义记录。
func (r *FlowRepo) CreateProcDef(ctx context.Context, procDef *entity.ProcDef) error {
	return r.ctxDB(ctx).Create(&procDef).Error
}

// ===================== 流程实例 =====================

// CreateInstance 创建流程实例记录。
func (r *FlowRepo) CreateInstance(ctx context.Context, inst *entity.ProcInst) error {
	return r.ctxDB(ctx).Create(&inst).Error
}

// UpdateInstance 更新流程实例指定字段。
func (r *FlowRepo) UpdateInstance(ctx context.Context, id int, updates map[string]any) error {
	return r.ctxDB(ctx).Model(&entity.ProcInst{}).Where("id=?", id).Updates(updates).Error
}

// GetInstanceInfo 通过 CTE 合并当前表和历史表获取流程实例信息。
func (r *FlowRepo) GetInstanceInfo(ctx context.Context, instID int) (model.InstanceView, error) {
	var inst model.InstanceView
	sql := "WITH tmp_procinst AS (" +
		"SELECT id, proc_id, proc_version, business_id, starter, current_node_id, " +
		"create_time, `status` FROM proc_inst WHERE id = ? " +
		"UNION ALL " +
		"SELECT proc_inst_id AS id, proc_id, proc_version, business_id, starter, current_node_id, " +
		"create_time, `status` FROM hist_proc_inst WHERE proc_inst_id = ?" +
		") " +
		"SELECT a.id, a.proc_id, a.proc_version, a.business_id, a.starter, " +
		"a.current_node_id, a.create_time, a.`status`, b.name " +
		"FROM tmp_procinst a " +
		"LEFT JOIN proc_def b ON a.proc_id = b.id"
	if err := r.ctxDB(ctx).Raw(sql, instID, instID).Scan(&inst).Error; err != nil {
		return inst, err
	}
	return inst, nil
}

// ListInstanceStartByUser 获取特定用户发起的流程实例列表（含分页）。
func (r *FlowRepo) ListInstanceStartByUser(ctx context.Context, p ListInstByUserParams) ([]model.InstanceView, error) {
	instFilter := "1=1"
	args := []any{}
	if p.UserID != "" {
		instFilter += " AND starter = ?"
		args = append(args, p.UserID)
	}

	sql := "WITH tmp_procinst AS (" +
		"SELECT id, proc_id, proc_version, business_id, starter, current_node_id, " +
		"create_time, `status` FROM proc_inst WHERE " + instFilter +
		" UNION ALL " +
		"SELECT proc_inst_id AS id, proc_id, proc_version, business_id, starter, current_node_id, " +
		"create_time, `status` FROM hist_proc_inst WHERE " + instFilter +
		") " +
		"SELECT a.id, a.proc_id, a.proc_version, a.business_id, " +
		"a.starter, a.current_node_id, a.create_time, a.`status`, b.name " +
		"FROM tmp_procinst a " +
		"JOIN proc_def b ON a.proc_id = b.id"

	// CTE 中 UNION ALL 的两段各用一次参数
	allArgs := append(append([]any{}, args...), args...)

	if p.ProcessName != "" {
		sql += " WHERE b.name = ?"
		allArgs = append(allArgs, p.ProcessName)
	}
	sql += " ORDER BY a.id LIMIT ? OFFSET ?"
	allArgs = append(allArgs, p.Limit, p.Offset)

	var instances []model.InstanceView
	if err := r.ctxDB(ctx).Raw(sql, allArgs...).Scan(&instances).Error; err != nil {
		return nil, err
	}
	return instances, nil
}

// CountInstanceStartByUser 获取特定用户发起的流程实例总数。
func (r *FlowRepo) CountInstanceStartByUser(ctx context.Context, p CountByUserParams) (int64, error) {
	instFilter := "1=1"
	args := []any{}
	if p.UserID != "" {
		instFilter += " AND starter = ?"
		args = append(args, p.UserID)
	}

	sql := "WITH tmp_procinst AS (" +
		"SELECT id FROM proc_inst WHERE " + instFilter +
		" UNION ALL " +
		"SELECT proc_inst_id AS id FROM hist_proc_inst WHERE " + instFilter +
		") " +
		"SELECT COUNT(*) " +
		"FROM tmp_procinst a " +
		"JOIN proc_def b ON a.proc_id = b.id"

	allArgs := append(append([]any{}, args...), args...)

	if p.ProcessName != "" {
		sql += " WHERE b.name = ?"
		allArgs = append(allArgs, p.ProcessName)
	}

	var count int64
	if err := r.ctxDB(ctx).Raw(sql, allArgs...).Scan(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetProcessIDByInstID 根据流程实例ID获取流程ID。
func (r *FlowRepo) GetProcessIDByInstID(ctx context.Context, instID int) (int, error) {
	var id int
	if err := r.ctxDB(ctx).
		Select("proc_id").
		Table("proc_inst").
		Where("id = ?", instID).
		Scan(&id).Error; err != nil {
		return 0, err
	}
	return id, nil
}

// GetProcessNameByInstID 根据流程实例ID获取流程名称。
func (r *FlowRepo) GetProcessNameByInstID(ctx context.Context, instID int) (string, error) {
	type res struct {
		Name string
	}
	var r2 res
	if err := r.ctxDB(ctx).
		Select("b.name").
		Table("proc_inst a").
		Joins("JOIN proc_def b ON a.proc_id = b.id").
		Where("a.id = ?", instID).
		Scan(&r2).Error; err != nil {
		return "", err
	}
	return r2.Name, nil
}

// ===================== 任务 =====================

// CreateTasks 批量创建任务记录。
func (r *FlowRepo) CreateTasks(ctx context.Context, tasks []entity.ProcTask) error {
	return r.ctxDB(ctx).Create(&tasks).Error
}

// UpdateTask 更新任务指定字段。
func (r *FlowRepo) UpdateTask(ctx context.Context, id int, updates map[string]any) error {
	return r.ctxDB(ctx).Model(&entity.ProcTask{}).Where("id=?", id).Updates(updates).Error
}

// GetTaskInfo 通过 CTE 合并当前表和历史表获取任务信息。
func (r *FlowRepo) GetTaskInfo(ctx context.Context, taskID int) (model.TaskView, error) {
	var task model.TaskView
	sql := "WITH tmp_task AS (" +
		"SELECT id, proc_id, proc_inst_id, business_id, starter, node_id, node_name, prev_node_id, " +
		"is_cosigned, batch_code, user_id, `status`, is_finished, `comment`, " +
		"proc_inst_create_time, create_time, finished_time FROM proc_task WHERE id = ? " +
		"UNION ALL " +
		"SELECT task_id AS id, proc_id, proc_inst_id, business_id, starter, node_id, node_name, " +
		"prev_node_id, is_cosigned, batch_code, user_id, `status`, is_finished, `comment`, " +
		"proc_inst_create_time, create_time, finished_time FROM hist_proc_task WHERE id = ?" +
		") " +
		"SELECT a.id, a.proc_id, b.name, a.proc_inst_id, a.business_id, a.starter, " +
		"a.node_id, a.node_name, a.prev_node_id, a.is_cosigned, " +
		"a.batch_code, a.user_id, a.`status`, a.is_finished, a.`comment`, " +
		"a.proc_inst_create_time, a.create_time, a.finished_time " +
		"FROM tmp_task a " +
		"LEFT JOIN proc_def b ON a.proc_id = b.id"
	if err := r.ctxDB(ctx).Raw(sql, taskID, taskID).Scan(&task).Error; err != nil {
		return model.TaskView{}, err
	}
	if task.TaskID == 0 {
		return model.TaskView{}, fmt.Errorf("ID为%d的任务不存在", taskID)
	}
	return task, nil
}

// taskSelectColumns 返回任务查询的 SELECT 列列表。
func taskSelectColumns() string {
	return "a.id, a.proc_id, b.name, a.proc_inst_id, a.business_id, a.starter, " +
		"a.node_id, a.node_name, a.prev_node_id, a.is_cosigned, a.batch_code, " +
		"a.user_id, a.`status`, a.is_finished, a.`comment`, " +
		"a.proc_inst_create_time, a.create_time, a.finished_time"
}

// ListTaskToDo 获取特定用户待办任务列表。
func (r *FlowRepo) ListTaskToDo(ctx context.Context, p ListToDoParams) ([]model.TaskView, error) {
	db := r.ctxDB(ctx).
		Table("proc_task a").
		Select(taskSelectColumns()).
		Joins("JOIN proc_def b ON a.proc_id = b.id").
		Where("a.is_finished = 0")

	if p.UserID != "" {
		db = db.Where("a.user_id = ?", p.UserID)
	}
	if p.ProcessName != "" {
		db = db.Where("b.name = ?", p.ProcessName)
	}

	if p.Asc {
		db = db.Order("a.id ASC")
	} else {
		db = db.Order("a.id DESC")
	}

	var tasks []model.TaskView
	if err := db.Offset(p.Offset).Limit(p.Limit).Scan(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// CountTaskToDo 获取特定用户待办任务总数。
func (r *FlowRepo) CountTaskToDo(ctx context.Context, p CountByUserParams) (int64, error) {
	db := r.ctxDB(ctx).
		Table("proc_task a").
		Joins("JOIN proc_def b ON a.proc_id = b.id").
		Where("a.is_finished = 0")

	if p.UserID != "" {
		db = db.Where("a.user_id = ?", p.UserID)
	}
	if p.ProcessName != "" {
		db = db.Where("b.name = ?", p.ProcessName)
	}

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ListTaskFinished 获取特定用户已完成任务列表。
func (r *FlowRepo) ListTaskFinished(ctx context.Context, p ListFinishedParams) ([]model.TaskView, error) {
	if p.UserID == "" {
		p.IgnoreStartByMe = false
	}

	taskFilter := "1=1"
	args := []any{}
	if p.UserID != "" {
		taskFilter += " AND user_id = ?"
		args = append(args, p.UserID)
	}

	sql := "WITH tmp_task AS (" +
		"SELECT id, proc_id, proc_inst_id, business_id, starter, node_id, node_name, prev_node_id, " +
		"is_cosigned, batch_code, user_id, `status`, is_finished, `comment`, " +
		"proc_inst_create_time, create_time, finished_time FROM proc_task WHERE " + taskFilter +
		" UNION ALL " +
		"SELECT task_id AS id, proc_id, proc_inst_id, business_id, starter, node_id, node_name, " +
		"prev_node_id, is_cosigned, batch_code, user_id, `status`, is_finished, `comment`, " +
		"proc_inst_create_time, create_time, finished_time FROM hist_proc_task WHERE " + taskFilter +
		") " +
		"SELECT a.id, a.proc_id, b.name, a.proc_inst_id, a.business_id, a.starter, " +
		"a.node_id, a.node_name, a.prev_node_id, a.is_cosigned, " +
		"a.batch_code, a.user_id, a.`status`, a.is_finished, a.`comment`, " +
		"a.proc_inst_create_time, a.create_time, a.finished_time " +
		"FROM tmp_task a " +
		"JOIN proc_def b ON a.proc_id = b.id " +
		"WHERE a.is_finished = 1 AND a.`status` != 0"

	// CTE 中 UNION ALL 的两段各用一次参数
	allArgs := append(append([]any{}, args...), args...)

	if p.ProcessName != "" {
		sql += " AND b.name = ?"
		allArgs = append(allArgs, p.ProcessName)
	}
	if p.IgnoreStartByMe {
		sql += " AND a.starter != ?"
		allArgs = append(allArgs, p.UserID)
	}

	if p.Asc {
		sql += " ORDER BY a.finished_time ASC"
	} else {
		sql += " ORDER BY a.finished_time DESC"
	}
	sql += " LIMIT ? OFFSET ?"
	allArgs = append(allArgs, p.Limit, p.Offset)

	var tasks []model.TaskView
	if err := r.ctxDB(ctx).Raw(sql, allArgs...).Scan(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// CountTaskFinished 获取特定用户已完成任务总数。
func (r *FlowRepo) CountTaskFinished(ctx context.Context, p CountFinishedParams) (int64, error) {
	if p.UserID == "" {
		p.IgnoreStartByMe = false
	}

	taskFilter := "1=1"
	args := []any{}
	if p.UserID != "" {
		taskFilter += " AND user_id = ?"
		args = append(args, p.UserID)
	}

	sql := "WITH tmp_task AS (" +
		"SELECT id, proc_id, proc_inst_id, business_id, starter, node_id, node_name, prev_node_id, " +
		"is_cosigned, batch_code, user_id, `status`, is_finished, `comment`, " +
		"proc_inst_create_time, create_time, finished_time FROM proc_task WHERE " + taskFilter +
		" UNION ALL " +
		"SELECT task_id AS id, proc_id, proc_inst_id, business_id, starter, node_id, node_name, " +
		"prev_node_id, is_cosigned, batch_code, user_id, `status`, is_finished, `comment`, " +
		"proc_inst_create_time, create_time, finished_time FROM hist_proc_task WHERE " + taskFilter +
		") " +
		"SELECT COUNT(*) " +
		"FROM tmp_task a " +
		"JOIN proc_def b ON a.proc_id = b.id " +
		"WHERE a.is_finished = 1 AND a.`status` != 0"

	allArgs := append(append([]any{}, args...), args...)

	if p.ProcessName != "" {
		sql += " AND b.name = ?"
		allArgs = append(allArgs, p.ProcessName)
	}
	if p.IgnoreStartByMe {
		sql += " AND a.starter != ?"
		allArgs = append(allArgs, p.UserID)
	}

	var count int64
	if err := r.ctxDB(ctx).Raw(sql, allArgs...).Scan(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ListInstanceTaskHistory 获取流程实例下所有任务历史记录。
func (r *FlowRepo) ListInstanceTaskHistory(ctx context.Context, instID int) ([]model.TaskView, error) {
	sql := "WITH tmp_task AS (" +
		"SELECT id, proc_id, proc_inst_id, business_id, starter, node_id, node_name, prev_node_id, " +
		"is_cosigned, batch_code, user_id, `status`, is_finished, `comment`, " +
		"proc_inst_create_time, create_time, finished_time FROM proc_task WHERE proc_inst_id = ? " +
		"UNION ALL " +
		"SELECT task_id AS id, proc_id, proc_inst_id, business_id, starter, node_id, node_name, " +
		"prev_node_id, is_cosigned, batch_code, user_id, `status`, is_finished, `comment`, " +
		"proc_inst_create_time, create_time, finished_time FROM hist_proc_task WHERE proc_inst_id = ?" +
		") " +
		"SELECT a.id, a.proc_id, b.name, a.proc_inst_id, a.business_id, a.starter, " +
		"a.node_id, a.node_name, a.prev_node_id, a.is_cosigned, " +
		"a.batch_code, a.user_id, a.`status`, a.is_finished, a.`comment`, " +
		"a.proc_inst_create_time, a.create_time, a.finished_time " +
		"FROM tmp_task a " +
		"JOIN proc_def b ON a.proc_id = b.id " +
		"ORDER BY a.id"
	var tasks []model.TaskView
	if err := r.ctxDB(ctx).Raw(sql, instID, instID).Scan(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTaskNodeStatus 获取任务节点审批状态，返回总任务数、通过数、驳回数。
func (r *FlowRepo) GetTaskNodeStatus(ctx context.Context, p TaskNodeStatusParams) (int, int, int, error) {
	type res struct {
		TotalTask     int `gorm:"column:total_task"`
		TotalPassed   int `gorm:"column:total_passed"`
		TotalRejected int `gorm:"column:total_rejected"`
	}
	var result res
	if err := r.ctxDB(ctx).
		Model(&entity.ProcTask{}).
		Select(
			"COUNT(*) AS total_task",
			"SUM(CASE `status` WHEN 1 THEN 1 ELSE 0 END) AS total_passed",
			"SUM(CASE `status` WHEN 2 THEN 1 ELSE 0 END) AS total_rejected",
		).
		Where("proc_inst_id = ? AND node_id = ? AND batch_code = ?", p.InstID, p.NodeID, p.BatchCode).
		Scan(&result).Error; err != nil {
		return 0, 0, 0, err
	}
	return result.TotalTask, result.TotalPassed, result.TotalRejected, nil
}

// GetNotFinishUsers 获取某节点中未完成任务的用户ID列表。
func (r *FlowRepo) GetNotFinishUsers(ctx context.Context, p NodeQueryParams) ([]string, error) {
	var userIDs []string
	if err := r.ctxDB(ctx).
		Model(&entity.ProcTask{}).
		Distinct("user_id").
		Where("proc_inst_id = ? AND node_id = ? AND is_finished = 0", p.InstID, p.NodeID).
		Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	return userIDs, nil
}

// GetPrevNodeBatchCode 获取任务的上一个实际执行节点的批次码。
func (r *FlowRepo) GetPrevNodeBatchCode(ctx context.Context, taskID int) (string, error) {
	type batchCodeRes struct {
		BatchCode string `gorm:"column:batch_code"`
	}
	var bc batchCodeRes
	if err := r.ctxDB(ctx).
		Table("proc_task a").
		Select("a.batch_code").
		Joins("JOIN proc_task b ON a.node_id = b.prev_node_id AND a.proc_inst_id = b.proc_inst_id").
		Where("b.id = ?", taskID).
		Order("a.id DESC").
		Limit(1).
		Scan(&bc).Error; err != nil {
		return "", err
	}
	return bc.BatchCode, nil
}

// HasRejectInBatch 检查指定批次码中是否存在驳回状态的任务。
func (r *FlowRepo) HasRejectInBatch(ctx context.Context, batchCode string) (bool, error) {
	var id int
	if err := r.ctxDB(ctx).
		Select("id").
		Table("proc_task").
		Where("batch_code = ? AND `status` = 2", batchCode).
		Limit(1).
		Scan(&id).Error; err != nil {
		return false, err
	}
	return id != 0, nil
}

// UpdateTasksByBatchCode 按批次码批量更新任务字段。
func (r *FlowRepo) UpdateTasksByBatchCode(ctx context.Context, batchCode string, updates map[string]any) error {
	return r.ctxDB(ctx).Model(&entity.ProcTask{}).Where("batch_code=?", batchCode).Updates(updates).Error
}

// DeleteTasksByBatchCode 按批次码删除任务。
func (r *FlowRepo) DeleteTasksByBatchCode(ctx context.Context, batchCode string) error {
	return r.ctxDB(ctx).Where("batch_code=?", batchCode).Delete(&entity.ProcTask{}).Error
}

// DeleteTaskByID 按任务ID删除任务。
func (r *FlowRepo) DeleteTaskByID(ctx context.Context, taskID int) error {
	return r.ctxDB(ctx).Where("id = ?", taskID).Delete(&entity.ProcTask{}).Error
}

// RevokeTask 将任务状态重置为初始（回滚已提交的任务）。
func (r *FlowRepo) RevokeTask(ctx context.Context, taskID int) error {
	db := r.ctxDB(ctx)
	// 首先将 task 状态做初始化
	if err := db.Model(&entity.ProcTask{}).
		Where("id = ?", taskID).
		Updates(map[string]any{
			"`status`":      0,
			"is_finished":   0,
			"finished_time": nil,
			"comment":       nil,
		}).Error; err != nil {
		return err
	}
	// 对那些没有做通过或驳回、被自动设置为 finish 的 task 做初始化
	// 先查询 batch_code，再用它做条件更新
	var batchCode string
	if err := db.Model(&entity.ProcTask{}).
		Select("batch_code").
		Where("id = ?", taskID).
		Scan(&batchCode).Error; err != nil {
		return err
	}
	return db.Model(&entity.ProcTask{}).
		Where("`status` = 0 AND batch_code = ?", batchCode).
		Update("is_finished", 0).Error
}

// GetNextNodeIDByPrevNodeID 获取流程定义中某节点的下一个节点ID。
func (r *FlowRepo) GetNextNodeIDByPrevNodeID(ctx context.Context, prevNodeID string) (string, error) {
	var execution entity.ProcExecution
	if err := r.ctxDB(ctx).Where("prev_node_id=?", prevNodeID).First(&execution).Error; err != nil {
		return "", err
	}
	return execution.NodeID, nil
}

// GetUpstreamNodes 递归获取任务所在节点的所有上游任务节点。
func (r *FlowRepo) GetUpstreamNodes(ctx context.Context, nodeID string) ([]model.Node, error) {
	sql := "WITH RECURSIVE tmp(node_id, node_name, prev_node_id, node_type) AS (" +
		"SELECT node_id, node_name, prev_node_id, node_type FROM proc_execution WHERE node_id = ? " +
		"UNION ALL " +
		"SELECT a.node_id, a.node_name, a.prev_node_id, a.node_type " +
		"FROM proc_execution a " +
		"JOIN tmp b ON a.node_id = b.prev_node_id" +
		") " +
		"SELECT DISTINCT node_id, node_name, prev_node_id, node_type " +
		"FROM tmp " +
		"WHERE node_type != 2 AND node_id != ?"
	var nodes []model.Node
	if err := r.ctxDB(ctx).Raw(sql, nodeID, nodeID).Scan(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

// ===================== 执行关系 =====================

// GetStartNodeID 获取流程的开始节点ID。
func (r *FlowRepo) GetStartNodeID(ctx context.Context, procID int) (string, error) {
	type res struct {
		NodeID string `gorm:"column:node_id"`
	}
	var r2 res
	if err := r.ctxDB(ctx).
		Select("node_id").
		Table("proc_execution").
		Where("proc_id = ? AND node_type = 0", procID).
		Scan(&r2).Error; err != nil {
		return "", err
	}
	return r2.NodeID, nil
}

// IsNodeFinished 判断特定实例中某一个节点是否已经全部完成。
func (r *FlowRepo) IsNodeFinished(ctx context.Context, p NodeQueryParams) (bool, error) {
	type countResult struct {
		Total    int `gorm:"column:total"`
		Finished int `gorm:"column:finished"`
	}
	var result countResult
	if err := r.ctxDB(ctx).
		Model(&entity.ProcTask{}).
		Select("COUNT(*) AS total, SUM(is_finished) AS finished").
		Where("proc_inst_id = ? AND node_id = ?", p.InstID, p.NodeID).
		Group("proc_inst_id, node_id").
		Scan(&result).Error; err != nil {
		return false, err
	}
	// 没有任何记录，认为已完成
	if result.Total == 0 {
		return true, nil
	}
	return result.Total == result.Finished, nil
}

// ===================== 变量 =====================

// SaveVariable 批量保存流程实例变量。已存在则更新，不存在则插入。
func (r *FlowRepo) SaveVariable(ctx context.Context, instID int, variables []model.Variable) error {
	if len(variables) == 0 {
		return nil
	}
	db := r.ctxDB(ctx)
	for _, v := range variables {
		var existing entity.ProcInstVariable
		if err := db.
			Where("proc_inst_id = ? AND `key` = ?", instID, v.Key).
			First(&existing).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if existing.ID == 0 {
			if err := db.Create(&entity.ProcInstVariable{ProcInstID: instID, Key: v.Key, Value: v.Value}).Error; err != nil {
				return err
			}
		} else {
			if err := db.Model(&entity.ProcInstVariable{}).
				Where("proc_inst_id = ? AND `key` = ?", instID, v.Key).
				Update("value", v.Value).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// GetVariable 获取流程实例变量值。返回变量值、是否存在、错误。
func (r *FlowRepo) GetVariable(ctx context.Context, instID int, key string) (string, bool, error) {
	type res struct {
		Value string
	}
	var r2 res
	if err := r.ctxDB(ctx).
		Select("`value`").
		Table("proc_inst_variable").
		Where("proc_inst_id = ? AND `key` = ?", instID, key).
		Limit(1).
		Scan(&r2).Error; err != nil {
		return "", false, err
	}
	if r2.Value != "" {
		return r2.Value, true, nil
	}
	return r2.Value, false, nil
}

// ResolveVariables 批量解析变量，返回变量名到值的映射。
// 非变量（不以$开头）的值原样存储在映射中。
func (r *FlowRepo) ResolveVariables(ctx context.Context, instID int, varNames []string) (map[string]string, error) {
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
func (r *FlowRepo) ArchiveInstance(ctx context.Context, instID int, status int) error {
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
