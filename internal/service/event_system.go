package service

import (
	"context"
	"fmt"

	"github.com/Bunny3th/easy-workflow/internal/event"
	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/pkg"
)

// RegisterNodeEvent 注册节点事件处理器。
// 节点事件用于 NodeStartEvents、NodeEndEvents、TaskFinishEvents。
func (e *Engine) RegisterNodeEvent(name string, handler event.NodeEventHandler) {
	e.eventMu.Lock()
	e.nodeEventPool[name] = handler
	e.eventMu.Unlock()
}

// RegisterProcEvent 注册流程事件处理器。
// 流程事件用于 RevokeEvents（流程撤销事件）。
func (e *Engine) RegisterProcEvent(name string, handler event.ProcEventHandler) {
	e.eventMu.Lock()
	e.procEventPool[name] = handler
	e.eventMu.Unlock()
}

// verifyEvents 检查流程的所有事件是否已注册。
func (e *Engine) verifyEvents(ctx context.Context, procID int, nodes map[string]model.Node) error {
	process, err := e.GetProcessDefine(ctx, procID)
	if err != nil {
		return err
	}

	// 验证流程撤销事件
	for _, name := range process.RevokeEvents {
		e.eventMu.RLock()
		_, ok := e.procEventPool[name]
		e.eventMu.RUnlock()
		if !ok {
			return fmt.Errorf("事件%s尚未注册", name)
		}
	}

	// 收集所有节点事件并去重
	var nodeEvents []string
	for _, node := range nodes {
		nodeEvents = append(nodeEvents, node.NodeStartEvents...)
		nodeEvents = append(nodeEvents, node.NodeEndEvents...)
		nodeEvents = append(nodeEvents, node.TaskFinishEvents...)
	}
	nodeEventsSet := pkg.MakeUnique(nodeEvents)

	// 验证节点事件
	for _, name := range nodeEventsSet {
		e.eventMu.RLock()
		_, ok := e.nodeEventPool[name]
		e.eventMu.RUnlock()
		if !ok {
			return fmt.Errorf("事件%s尚未注册", name)
		}
	}

	return nil
}

// runNodeEvents 运行节点事件（节点开始、节点结束、任务完成）。
func (e *Engine) runNodeEvents(ctx context.Context, eventNames []string, id int, currentNode *model.Node, prevNode model.Node) error {
	for _, name := range eventNames {
		e.eventMu.RLock()
		handler, ok := e.nodeEventPool[name]
		e.eventMu.RUnlock()
		if !ok {
			return fmt.Errorf("事件%s未注册", name)
		}

		if err := handler(ctx, id, currentNode, prevNode); err != nil && !e.ignoreEventErr {
			return fmt.Errorf("节点[%s]事件[%s]执行出错:%v", currentNode.NodeName, name, err)
		}
	}
	return nil
}

// runProcEvents 运行流程事件（目前只有撤销事件）。
func (e *Engine) runProcEvents(ctx context.Context, eventNames []string, instID int, revokeUserID string) error {
	processName, err := e.repo.GetProcessNameByInstID(ctx, instID)
	if err != nil {
		return err
	}

	for _, name := range eventNames {
		e.eventMu.RLock()
		handler, ok := e.procEventPool[name]
		e.eventMu.RUnlock()
		if !ok {
			return fmt.Errorf("事件%s未注册", name)
		}

		if err := handler(ctx, instID, revokeUserID); err != nil && !e.ignoreEventErr {
			return fmt.Errorf("流程[%s]撤销事件[%s]执行出错:%v", processName, name, err)
		}
	}
	return nil
}
