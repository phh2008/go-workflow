package model

// InstanceStartParams 启动流程实例参数。
type InstanceStartParams struct {
	ProcessID     int    // 流程ID
	BusinessID    string // 业务ID
	Comment       string // 评论意见
	VariablesJSON string // 变量 JSON
}

// InstanceRevokeParams 撤销流程实例参数。
type InstanceRevokeParams struct {
	InstanceID   int    // 流程实例ID
	Force        bool   // 是否强制撤销
	RevokeUserID string // 撤销发起用户ID
}

// TaskPassParams 任务通过参数。
type TaskPassParams struct {
	TaskID             int    // 任务ID
	Comment            string // 评论意见
	VariableJSON       string // 变量 JSON
	DirectlyToRejected bool   // 是否直接跳到上一个驳回自己的节点
}

// TaskRejectParams 任务驳回参数。
type TaskRejectParams struct {
	TaskID       int    // 任务ID
	Comment      string // 评论意见
	VariableJSON string // 变量 JSON
}

// TaskFreeRejectParams 自由驳回参数。
type TaskFreeRejectParams struct {
	TaskID         int    // 任务ID
	RejectToNodeID string // 驳回到的目标节点ID
	Comment        string // 评论意见
	VariableJSON   string // 变量 JSON
}

// TaskTransferParams 任务转交参数。
type TaskTransferParams struct {
	TaskID int      // 任务ID
	Users  []string // 目标用户ID列表
}

// TaskToDoListParams 待办任务列表查询参数。
type TaskToDoListParams struct {
	UserID      string // 用户ID，空则查询所有用户
	ProcessName string // 流程名称，空则查询所有流程
	Asc         bool   // 是否升序排列
	PageNo      int    // 页码
	PageSize    int    // 每页数量
}

// TaskFinishedListParams 已办任务列表查询参数。
type TaskFinishedListParams struct {
	UserID          string // 用户ID，空则查询所有用户
	ProcessName     string // 流程名称，空则查询所有流程
	IgnoreStartByMe bool   // 是否忽略自己发起流程中自己是处理人的任务
	Asc             bool   // 是否升序排列
	PageNo          int    // 页码
	PageSize        int    // 每页数量
}

// InstanceListByUserParams 用户发起的流程实例列表查询参数。
type InstanceListByUserParams struct {
	UserID      string // 用户ID，空则查询所有用户
	ProcessName string // 流程名称，空则查询所有流程
	PageNo      int    // 页码
	PageSize    int    // 每页数量
}

// ProcessSaveParams 保存流程定义参数。
type ProcessSaveParams struct {
	Resource     string // 流程定义 JSON
	CreateUserID string // 创建者ID
}

// ResolveVariablesParams 解析变量参数。
type ResolveVariablesParams struct {
	InstanceID int      // 流程实例ID
	Variables  []string // 变量名列表
}
