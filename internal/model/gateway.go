package model

// HybridGateway 混合网关，等同于 Activiti 中排他网关、并行网关、包含网关的混合体。
type HybridGateway struct {
	Conditions      []Condition // 条件判断节点列表
	InevitableNodes []string    // 必然执行的节点ID列表
	WaitForAllPrevNode int       // 等待模式：0 包含网关（任一上级完成即可），1 并行网关（全部上级完成才行）
}

// Condition 表示网关中的条件分支。
type Condition struct {
	Expression string // 条件表达式，使用 $变量名 引用流程变量
	NodeID     string // 满足条件后跳转到的目标节点ID
}
