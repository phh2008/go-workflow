package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/pkg"
	"github.com/Bunny3th/easy-workflow/internal/repository"
	"gorm.io/gorm"
)

// processNode 处理节点，如：生成task、进行条件判断、处理结束节点等。
func (e *Engine) processNode(ctx context.Context, instID int, current *model.Node, prev model.Node) error {
	slog.Debug("[processNode] 开始处理节点", "instID", instID, "nodeID", current.NodeID, "nodeName", current.NodeName, "nodeType", current.NodeType)

	// 处理节点开始事件
	if err := e.runNodeEvents(ctx, current.NodeStartEvents, instID, current, prev); err != nil {
		return err
	}

	switch current.NodeType {
	case model.RootNode:
		_, err := e.taskNodeHandle(ctx, instID, current, prev)
		return err
	case model.GateWayNode:
		return e.gatewayNodeHandle(ctx, instID, current, prev)
	case model.TaskNode:
		_, err := e.taskNodeHandle(ctx, instID, current, prev)
		return err
	case model.EndNode:
		return e.endNodeHandle(ctx, instID, 1)
	}

	return nil
}

// taskNodeHandle 任务节点处理，返回生成的 taskID 数组。
func (e *Engine) taskNodeHandle(ctx context.Context, instID int, current *model.Node, prev model.Node) ([]int, error) {
	// 获取节点用户
	users, err := e.resolveNodeUsers(ctx, instID, *current)
	if err != nil {
		return nil, err
	}

	if len(users) < 1 {
		return nil, fmt.Errorf("未指定处理人，无法处理节点:%s", current.NodeName)
	}

	// 开始节点只能有一个用户发起
	if current.NodeType == model.RootNode {
		users = users[0:1]
	}

	// 生成Task
	return e.createTask(ctx, instID, current.NodeID, prev.NodeID, users)
}

// gatewayNodeHandle 网关节点处理。
func (e *Engine) gatewayNodeHandle(ctx context.Context, instID int, current *model.Node, prevTaskNode model.Node) error {
	// 确认上级节点是否已全部完成
	totalFinished := 0
	totalPrevNodes := len(current.PrevNodeIDs)
	for _, nodeID := range current.PrevNodeIDs {
		finished, err := e.repo.IsNodeFinished(ctx, repository.NodeQueryParams{InstID: instID, NodeID: nodeID})
		if err != nil {
			return err
		}
		slog.Debug("[gatewayNodeHandle] 检查上级节点完成情况", "instID", instID, "nodeID", current.NodeID, "prevNodeID", nodeID, "finished", finished)
		if finished {
			totalFinished++
		}
	}
	slog.Debug("[gatewayNodeHandle] 上级节点统计", "instID", instID, "nodeID", current.NodeID, "totalPrevNodes", totalPrevNodes, "totalFinished", totalFinished, "waitForAll", current.GWConfig.WaitForAllPrevNode)

	// 并行网关模式：还有尚未完成的上级节点，则退出
	if current.GWConfig.WaitForAllPrevNode == 1 && totalPrevNodes != totalFinished {
		slog.Debug("[gatewayNodeHandle] 并行网关，上级节点未全部完成，退出", "instID", instID, "nodeID", current.NodeID)
		return nil
	}

	// 包含网关模式：连一个已完成的上级节点都没有，则退出
	if current.GWConfig.WaitForAllPrevNode == 0 && totalFinished < 1 {
		slog.Debug("[gatewayNodeHandle] 包含网关，无上级节点完成，退出", "instID", instID, "nodeID", current.NodeID)
		return nil
	}

	// 计算条件
	var conditionNodeIDs []string
	for _, c := range current.GWConfig.Conditions {
		reg := regexp.MustCompile(`[$]\w+`)
		variables := reg.FindAllString(c.Expression, -1)

		// 替换表达式中的变量为值
		expression := c.Expression
		kv, err := e.ResolveVariables(ctx, model.ResolveVariablesParams{InstanceID: instID, Variables: variables})
		if err != nil {
			return err
		}
		for k, v := range kv {
			expression = strings.ReplaceAll(expression, k, v)
		}

		// 使用表达式求值器计算表达式
		ok, err := e.expressionEval.EvalWithRawEnv(expression, kv)
		if err != nil {
			slog.Debug("[gatewayNodeHandle] 表达式求值失败", "expression", expression, "kv", kv, "error", err)
			return err
		}
		slog.Debug("[gatewayNodeHandle] 表达式求值结果", "instID", instID, "nodeID", current.NodeID, "rawExpression", c.Expression, "resolvedExpression", expression, "kv", kv, "result", ok)
		if ok {
			conditionNodeIDs = append(conditionNodeIDs, c.NodeID)
		}
	}

	// 将 conditionNodeIDs 和 InevitableNodes 合并去重
	nextNodeIDs := pkg.MakeUnique(conditionNodeIDs, current.GWConfig.InevitableNodes)
	slog.Debug("[gatewayNodeHandle] 下级节点列表", "instID", instID, "nodeID", current.NodeID, "conditionNodes", conditionNodeIDs, "inevitableNodes", current.GWConfig.InevitableNodes, "nextNodeIDs", nextNodeIDs)

	// 处理节点结束事件
	if err := e.runNodeEvents(ctx, current.NodeEndEvents, instID, current, prevTaskNode); err != nil {
		return err
	}

	// 对下级节点进行处理
	for _, nodeID := range nextNodeIDs {
		nextNode, err := e.getInstanceNode(ctx, instID, nodeID)
		if err != nil {
			return err
		}
		err = e.processNode(ctx, instID, &nextNode, prevTaskNode)
		if err != nil {
			return err
		}
	}

	return nil
}

// endNodeHandle 结束节点处理，将数据库中此流程实例产生的数据归档。
// status 流程实例状态：1 已完成，2 撤销。
func (e *Engine) endNodeHandle(ctx context.Context, instID int, status int) error {
	slog.Debug("[endNodeHandle] 流程结束，开始归档", "instID", instID, "status", status)
	return e.db.Transaction(func(tx *gorm.DB) error {
		txCtx := repository.WithTx(ctx, tx)
		return e.repo.ArchiveInstance(txCtx, instID, status)
	})
}

// getInstanceNode 获取流程实例中某个节点。
func (e *Engine) getInstanceNode(ctx context.Context, instID int, nodeID string) (model.Node, error) {
	procID, err := e.repo.GetProcessIDByInstID(ctx, instID)
	if err != nil {
		return model.Node{}, err
	}

	nodes, err := e.getProcCache(ctx, procID)
	if err != nil {
		return model.Node{}, err
	}

	node, ok := nodes[nodeID]
	if !ok {
		return model.Node{}, fmt.Errorf("ID为%d的流程实例中不存在ID为%s的节点", instID, nodeID)
	}

	return node, nil
}

// resolveNodeUsers 解析节点用户：获得用户变量并去重
func (e *Engine) resolveNodeUsers(ctx context.Context, instID int, node model.Node) ([]string, error) {
	kv, err := e.ResolveVariables(ctx, model.ResolveVariablesParams{InstanceID: instID, Variables: node.UserIDs})
	if err != nil {
		return nil, err
	}
	// 使用map去重
	usersMap := make(map[string]struct{})
	for _, v := range kv {
		usersMap[v] = struct{}{}
	}
	var users []string
	for k := range usersMap {
		users = append(users, k)
	}
	return users, nil
}

// resolveNodeUsersFromVars 解析节点用户：获得用户变量并去重
func (e *Engine) resolveNodeUsersFromVars(ctx context.Context, node model.Node, vars map[string]string) ([]string, error) {
	_ = ctx
	result := make(map[string]string)
	for _, v := range node.UserIDs {
		if after, ok := strings.CutPrefix(v, "$"); ok {
			key := after
			value, ok := vars[key]
			if !ok {
				return nil, errors.New("无法匹配变量:" + v)
			}
			result[v] = value
		} else {
			result[v] = v
		}
	}
	// 使用map去重
	usersMap := make(map[string]struct{})
	for _, v := range result {
		usersMap[v] = struct{}{}
	}
	var users []string
	for k := range usersMap {
		users = append(users, k)
	}
	return users, nil
}
