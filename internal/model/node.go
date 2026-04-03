package model

// NodeType 表示流程节点的类型。
type NodeType int

const (
	// RootNode 开始（根）节点。
	RootNode NodeType = 0
	// TaskNode 任务节点，需要人处理的审批节点。
	TaskNode NodeType = 1
	// GateWayNode 网关节点，使用混合网关，等同于 Activiti 中排他网关、并行网关、包含网关的混合体。
	GateWayNode NodeType = 2
	// EndNode 结束节点，流程到达此节点则流程实例完成。
	EndNode NodeType = 3
)

// Node 表示流程中的一个节点。
type Node struct {
	NodeID           string        // 节点ID
	NodeName         string        // 节点名称
	NodeType         NodeType      // 节点类型：0 开始节点，1 任务节点，2 网关，3 结束节点
	PrevNodeIDs      []string      // 上级节点ID列表（因分支的存在，上级节点可能有多个）
	UserIDs          []string      // 节点处理人ID数组
	Roles            []string      // 节点处理角色数组（需通过 NodeStartEvents 解析角色为用户ID加入 UserIDs）
	GWConfig         HybridGateway // 网关配置，仅节点类型为 GateWay 时有值
	IsCosigned       int           // 是否会签：0 非会签，1 会签（全部通过才通过，一人驳回即驳回）
	NodeStartEvents  []string      // 节点开始时触发的事件方法名列表
	NodeEndEvents    []string      // 节点结束时触发的事件方法名列表
	TaskFinishEvents []string      // 任务完成（通过或驳回）时触发的事件方法名列表
}

// 设计思想说明：
//
// 一、开始节点与结束节点：
//  1. 遵循 BPMN 2.0 标准，必须包含开始和结束节点。
//  2. 与 Activiti 不同，本引擎中的开始节点同时是一个特殊的任务节点，
//     流程一旦启动，开始节点会自动通过，无需额外编写事件来自动结束。
//  3. 明确设置结束节点是一种防呆设计，避免分支较多时因疏忽导致流程卡死。
//
// 二、关于事件：
//  1. 结束节点只处理 NodeStartEvents，不处理 NodeEndEvents（结束节点仅做数据归档）。
//  2. 不推荐在网关节点上添加事件，会导致节点分配出错且可能多次触发。
//  3. 流程撤销事件属于整个流程，在流程定义时确定。
//
// 三、关于会签：
//  - 会签节点：全部通过才能通过，一人驳回即驳回。
//  - 非会签节点：一人通过即通过，一人驳回即驳回。
