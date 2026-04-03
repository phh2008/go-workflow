package model

// PageQuery 分页查询参数。
type PageQuery struct {
	PageNo   int `form:"pageNo" json:"pageNo"`
	PageSize int `form:"pageSize" json:"pageSize"`
}

// GetPageNo 返回页码，小于等于 0 时返回默认值 1。
func (p PageQuery) GetPageNo() int {
	if p.PageNo <= 0 {
		return 1
	}
	return p.PageNo
}

// GetPageSize 返回每页数量，小于等于 0 时返回默认值 10。
func (p PageQuery) GetPageSize() int {
	if p.PageSize <= 0 {
		return 10
	}
	return p.PageSize
}

// Offset 返回 SQL 分页偏移量。
func (p PageQuery) Offset() int {
	return (p.GetPageNo() - 1) * p.GetPageSize()
}

// PageData 泛型分页响应数据。
type PageData[T any] struct {
	Count    int64 `json:"count"`
	PageNo   int   `json:"pageNo"`
	PageSize int   `json:"pageSize"`
	Data     []T   `json:"data"`
}

// NewPageData 创建分页响应。
func NewPageData[T any](pageNo, pageSize int) *PageData[T] {
	return &PageData[T]{PageNo: pageNo, PageSize: pageSize}
}

// SetData 设置分页数据并返回自身，支持链式调用。
func (p *PageData[T]) SetData(data []T) *PageData[T] {
	p.Data = data
	return p
}

// SetCount 设置总记录数并返回自身，支持链式调用。
func (p *PageData[T]) SetCount(count int64) *PageData[T] {
	p.Count = count
	return p
}

// --- 流程定义请求 ---

// ProcessListReq 流程定义列表查询请求。
type ProcessListReq struct {
	Source string `form:"source" json:"source" binding:"required"`
}

// ProcessDefGetReq 获取流程定义请求。
type ProcessDefGetReq struct {
	ID int `form:"id" json:"id" binding:"required"`
}

// ProcessSaveReq 保存流程定义请求。
type ProcessSaveReq struct {
	Resource     string `form:"resource" json:"resource" binding:"required"`
	CreateUserID string `form:"createUserId" json:"createUserId" binding:"required"`
}

// --- 流程实例请求 ---

// InstanceStartReq 启动流程实例请求。
type InstanceStartReq struct {
	ProcessID     int    `form:"processId" json:"processId" binding:"required"`
	BusinessID    string `form:"businessId" json:"businessId" binding:"required"`
	Comment       string `form:"comment" json:"comment"`
	VariablesJSON string `form:"variablesJson" json:"variablesJson"`
}

// InstanceRevokeReq 撤销流程实例请求。
type InstanceRevokeReq struct {
	InstanceID   int    `form:"instanceId" json:"instanceId" binding:"required"`
	RevokeUserID string `form:"revokeUserId" json:"revokeUserId" binding:"required"`
	Force        bool   `form:"force" json:"force"`
}

// InstanceListReq 流程实例列表查询请求。
type InstanceListReq struct {
	PageQuery
	UserID      string `form:"userId" json:"userId"`
	ProcessName string `form:"processName" json:"processName"`
}

// InstanceTaskHistoryReq 流程实例任务历史请求。
type InstanceTaskHistoryReq struct {
	InstanceID int `form:"instid" json:"instid" binding:"required"`
}

// --- 任务请求 ---

// TaskActionReq 任务操作请求（Pass/Reject 共享）。
type TaskActionReq struct {
	TaskID       int    `form:"taskId" json:"taskId" binding:"required"`
	Comment      string `form:"comment" json:"comment"`
	VariableJSON string `form:"variableJson" json:"variableJson"`
}

// TaskFreeRejectReq 自由驳回请求。
type TaskFreeRejectReq struct {
	TaskActionReq
	RejectToNodeID string `form:"rejectToNodeId" json:"rejectToNodeId"`
}

// TaskTransferReq 任务转交请求。
type TaskTransferReq struct {
	TaskID int      `form:"taskId" json:"taskId" binding:"required"`
	Users  []string `form:"users" json:"users" binding:"required"`
}

// TaskListReq 任务列表查询请求（待办共享）。
type TaskListReq struct {
	PageQuery
	UserID      string `form:"userId" json:"userId"`
	ProcessName string `form:"processName" json:"processName"`
	Asc         bool   `form:"asc" json:"asc"`
}

// TaskFinishedListReq 已办任务查询请求。
type TaskFinishedListReq struct {
	TaskListReq
	IgnoreStartByMe bool `form:"ignoreStartByMe" json:"ignoreStartByMe"`
}

// TaskInfoReq 任务信息查询请求。
type TaskInfoReq struct {
	TaskID int `form:"taskid" json:"taskid" binding:"required"`
}
