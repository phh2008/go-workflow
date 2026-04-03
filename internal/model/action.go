package model

// TaskAction 表示任务可执行的操作集合。
type TaskAction struct {
	CanPass                     bool // 任务可以执行"通过"
	CanReject                   bool // 任务可以执行"驳回"
	CanFreeRejectToUpstreamNode bool // 任务可以执行"自由驳回"（驳回到任意上游节点）
	CanDirectlyToWhoRejectedMe  bool // 任务可以执行"直接提交到上一个驳回我的节点"
	CanRevoke                   bool // 任务可以执行"撤销"
}
