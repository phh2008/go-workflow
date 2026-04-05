package repository

// GetProcIDByNameParams 按名称查询流程ID参数。
type GetProcIDByNameParams struct {
	Name   string
	Source string
}

// UpdateProcDefParams 更新流程定义参数。
type UpdateProcDefParams struct {
	Name     string
	Source   string
	Resource string
	UserID   string
	Version  int
}

// CountByUserParams 按用户查询数量参数（共享）。
type CountByUserParams struct {
	UserID      string
	ProcessName string
}

// ListInstByUserParams 用户发起的实例列表查询参数。
type ListInstByUserParams struct {
	UserID      string
	ProcessName string
	Offset      int
	Limit       int
}

// ListToDoParams 待办任务列表查询参数。
type ListToDoParams struct {
	UserID      string
	ProcessName string
	Asc         bool
	Offset      int
	Limit       int
}

// CountFinishedParams 已办任务计数参数。
type CountFinishedParams struct {
	UserID          string
	ProcessName     string
	IgnoreStartByMe bool
}

// ListFinishedParams 已办任务列表查询参数。
type ListFinishedParams struct {
	UserID          string
	ProcessName     string
	IgnoreStartByMe bool
	Asc             bool
	Offset          int
	Limit           int
}

// TaskNodeStatusParams 任务节点状态查询参数。
type TaskNodeStatusParams struct {
	InstID    int
	NodeID    string
	BatchCode string
}

// NotFinishUsersParams 未完成任务用户查询参数。
type NotFinishUsersParams struct {
	InstID int
	NodeID string
}

// IsNodeFinishedParams 节点完成状态查询参数。
type IsNodeFinishedParams struct {
	InstID int
	NodeID string
}
