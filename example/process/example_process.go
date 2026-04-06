package process

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Bunny3th/easy-workflow"
	"github.com/Bunny3th/easy-workflow/internal/model"
)

// CreateProcessJson 定义一个示例流程，返回流程定义 JSON 字符串。
func CreateProcessJson() (string, error) {
	// 初始节点
	// 建议所有的初始节点User都定义为变量"$starter"，方便之后的开发管理
	Node1 := model.Node{NodeID: "Start", NodeName: "请假",
		NodeType: model.RootNode, UserIDs: []string{"$starter"},
		NodeEndEvents: []string{"MyEvent_End"},
	}

	// 网关,根据请假天数做判断:
	// 请假天数>=3,流程转到主管审批
	// 请假天数<3,流程结束
	GWConfig_Conditional := model.HybridGateway{Conditions: []model.Condition{
		{Expression: "$days>=3", NodeID: "Manager"},
		{Expression: "$days<3", NodeID: "END"},
	}, WaitForAllPrevNode: 0}
	Node2 := model.Node{NodeID: "GW-Day", NodeName: "请假天数判断",
		NodeType: model.GateWayNode, GWConfig: GWConfig_Conditional,
		PrevNodeIDs: []string{"Start"},
	}

	// 主管审批节点
	// 注意，这里使用了角色。因为系统无法预先知道角色中存在多少用户，所以必须用StartEvents解析角色，将角色中的用户加到UserIDs中
	Node3 := model.Node{NodeID: "Manager", NodeName: "主管审批",
		NodeType: model.TaskNode, Roles: []string{"主管"},
		PrevNodeIDs: []string{"GW-Day"},
		NodeStartEvents: []string{"MyEvent_ResolveRoles", "MyEvent_Notify"},
		NodeEndEvents:   []string{"MyEvent_End"},
	}

	// 网关,要求"主管审批节点"通过后，并行进入到"人事审批"以及"副总审批"节点
	GW_Parallel := model.HybridGateway{InevitableNodes: []string{"HR", "DeputyBoss"}, WaitForAllPrevNode: 0}
	Node4 := model.Node{NodeID: "GW-Parallel", NodeName: "并行网关",
		NodeType: model.GateWayNode, GWConfig: GW_Parallel,
		PrevNodeIDs: []string{"Manager"},
	}

	// 人事审批任务节点
	Node5 := model.Node{NodeID: "HR", NodeName: "人事审批",
		NodeType: model.TaskNode, Roles: []string{"人事经理"},
		PrevNodeIDs: []string{"GW-Parallel"},
		NodeStartEvents: []string{"MyEvent_ResolveRoles", "MyEvent_Notify"},
		NodeEndEvents:   []string{"MyEvent_End"},
	}

	// 副总审批任务节点
	// 注意，IsCosigned=1说明这是一个会签节点，全部通过才算通过，一人驳回即算驳回
	Node6 := model.Node{NodeID: "DeputyBoss", NodeName: "副总审批",
		NodeType: model.TaskNode, Roles: []string{"副总"},
		IsCosigned:  1,
		PrevNodeIDs: []string{"GW-Parallel"},
		NodeStartEvents: []string{"MyEvent_ResolveRoles", "MyEvent_Notify"},
		NodeEndEvents:   []string{"MyEvent_End"},
		TaskFinishEvents: []string{"MyEvent_TaskForceNodePass"},
	}

	// 此网关承接上一个NodeID=GW-Parallel的网关
	// WaitForAllPrevNode=1 等于并行网关，必须要上级节点"人事、副总"全部完成才能往下走
	GW_Parallel2 := model.HybridGateway{InevitableNodes: []string{"Boss"}, WaitForAllPrevNode: 1}
	Node7 := model.Node{NodeID: "GW-Parallel2", NodeName: "并行网关",
		NodeType:    model.GateWayNode,
		PrevNodeIDs: []string{"HR", "DeputyBoss"},
		GWConfig:    GW_Parallel2,
	}

	// 老板审批任务节点
	Node8 := model.Node{NodeID: "Boss", NodeName: "老板审批",
		NodeType: model.TaskNode, Roles: []string{"老板"},
		PrevNodeIDs: []string{"GW-Parallel2"},
		NodeStartEvents: []string{"MyEvent_ResolveRoles", "MyEvent_Notify"},
		NodeEndEvents:   []string{"MyEvent_End"},
	}

	// 结束节点
	Node9 := model.Node{NodeID: "END", NodeName: "END",
		NodeType: model.EndNode, PrevNodeIDs: []string{"GW-Day", "Boss"},
		NodeStartEvents: []string{"MyEvent_Notify"}}

	// 流程是节点的集合
	nodes := []model.Node{Node1, Node2, Node3, Node4, Node5, Node6, Node7, Node8, Node9}

	proc := model.Process{
		ProcessName:  "员工请假",
		Source:       "办公系统",
		RevokeEvents: []string{"MyEvent_Revoke"},
		Nodes:        nodes,
	}

	// 转化为json
	j, err := json.Marshal(proc)
	return string(j), err
}

// CreateExampleProcess 创建并保存示例流程。
func CreateExampleProcess(eng *easyworkflow.Engine) {
	// 获得示例流程json
	j, err := CreateProcessJson()
	if err != nil {
		slog.Error("生成流程JSON失败", "error", err)
		return
	}

	// 保存流程
	id, err := eng.ProcessSave(context.Background(), model.ProcessSaveReq{Resource: j, CreateUserID: "system"})
	if err != nil {
		slog.Error("保存流程失败", "error", err)
		return
	}
	slog.Info("流程保存成功", "流程ID", id)
}
