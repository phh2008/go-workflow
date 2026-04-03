package model

// Process 表示一个流程定义。
type Process struct {
	ProcessName  string   // 流程名称
	Source       string   // 来源（引擎可能被多个系统使用，记录流程的创建来源）
	RevokeEvents []string // 流程撤销事件，在流程实例撤销时触发
	Nodes        []Node   // 流程节点列表
}
