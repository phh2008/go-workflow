package service

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/pkg"
)

// RegisterEvents 注册一个或多个 struct 中的所有方法到事件池。
// 如果 struct 具有 SetEngine(*Engine) 方法，则自动注入引擎实例。
func (e *Engine) RegisterEvents(structs ...any) {
	for _, s := range structs {
		if s == nil {
			continue
		}

		structValue := reflect.ValueOf(s)
		structType := structValue.Type()

		for m := range structType.Methods() {
			m := m
			em := &eventMethod{receiver: s, method: m}
			e.eventPoolMu.Lock()
			e.eventPool[m.Name] = em
			e.eventPoolMu.Unlock()
		}

		// 自动注入 Engine：如果 struct 有 SetEngine(*Engine) 方法，则调用
		setEngineMethod := structValue.MethodByName("SetEngine")
		if setEngineMethod.IsValid() {
			setEngineMethod.Call([]reflect.Value{reflect.ValueOf(e)})
		}
	}
}

// verifyEvents 检查流程的所有事件是否已注册且参数正确。
func (e *Engine) verifyEvents(ctx context.Context, procID int, nodes map[string]model.Node) error {
	process, err := e.GetProcessDefine(ctx, procID)
	if err != nil {
		return err
	}

	// 验证流程撤销事件
	for _, event := range process.RevokeEvents {
		em, ok := e.getEvent(event)
		if !ok {
			return fmt.Errorf("事件%s尚未导入", event)
		}
		if err := e.verifyProcEventParams(em); err != nil {
			return err
		}
	}

	// 收集所有节点事件
	var nodeEvents []string
	for _, node := range nodes {
		nodeEvents = append(nodeEvents, node.NodeStartEvents...)
		nodeEvents = append(nodeEvents, node.NodeEndEvents...)
		nodeEvents = append(nodeEvents, node.TaskFinishEvents...)
	}

	// 去重
	nodeEventsSet := pkg.MakeUnique(nodeEvents)

	// 验证节点事件
	for _, event := range nodeEventsSet {
		em, ok := e.getEvent(event)
		if !ok {
			return fmt.Errorf("事件%s尚未导入", event)
		}
		if err := e.verifyNodeEventParams(em); err != nil {
			return err
		}
	}

	return nil
}

// runNodeEvents 运行节点事件（节点开始、节点结束、任务完成）。
func (e *Engine) runNodeEvents(ctx context.Context, eventNames []string, id int, currentNode *model.Node, prevNode model.Node) error {
	_ = ctx
	for _, name := range eventNames {
		em, ok := e.getEvent(name)
		if !ok {
			return fmt.Errorf("事件%s未注册", name)
		}

		m := em.method.(reflect.Method)
		arg := []reflect.Value{
			reflect.ValueOf(em.receiver),
			reflect.ValueOf(id),
			reflect.ValueOf(currentNode),
			reflect.ValueOf(prevNode),
		}

		result := m.Func.Call(arg)

		if !e.ignoreEventErr {
			if !result[0].IsNil() {
				return fmt.Errorf("节点[%s]事件[%s]执行出错:%v", currentNode.NodeName, m.Name, result[0])
			}
		}
	}
	return nil
}

// runProcEvents 运行流程事件（目前只有撤销事件）。
func (e *Engine) runProcEvents(ctx context.Context, eventNames []string, instID int, revokeUserID string) error {
	_ = ctx
	for _, name := range eventNames {
		em, ok := e.getEvent(name)
		if !ok {
			return fmt.Errorf("事件%s未注册", name)
		}

		m := em.method.(reflect.Method)

		processName, err := e.repo.GetProcessNameByInstID(ctx, instID)
		if err != nil {
			return err
		}

		arg := []reflect.Value{
			reflect.ValueOf(em.receiver),
			reflect.ValueOf(instID),
			reflect.ValueOf(revokeUserID),
		}

		result := m.Func.Call(arg)

		if !e.ignoreEventErr {
			if !result[0].IsNil() {
				return fmt.Errorf("流程[%s]撤销事件[%s]执行出错:%v", processName, m.Name, result[0])
			}
		}
	}
	return nil
}

// verifyProcEventParams 验证流程事件参数是否正确。
// 流程撤销事件签名：func(struct *interface{}, ProcessInstanceID int, RevokeUserID string) error
func (e *Engine) verifyProcEventParams(em *eventMethod) error {
	m := em.method.(reflect.Method)

	if m.Type.NumIn() != 3 || m.Type.NumOut() != 1 {
		return fmt.Errorf("warning:事件方法 %s 入参、出参数量不匹配,此函数无法运行", m.Name)
	}

	if m.Type.In(1).Kind().String() != "int" {
		return fmt.Errorf("warning:事件方法 %s 参数1不是int类型,此函数无法运行", m.Name)
	}

	if m.Type.In(2).Kind().String() != "string" {
		return fmt.Errorf("warning:事件方法 %s 参数2不是string类型,此函数无法运行", m.Name)
	}

	if !pkg.TypeIsError(m.Type.Out(0)) {
		return fmt.Errorf("warning:事件方法 %s 返回参数不是error类型,此函数无法运行", m.Name)
	}
	return nil
}

// verifyNodeEventParams 验证节点事件参数是否正确。
// 节点开始/结束事件签名：func(struct *interface{}, ProcessInstanceID int, CurrentNode *Node, PrevNode Node) error
// 任务完成事件签名：func(struct *interface{}, TaskID int, CurrentNode *Node, PrevNode Node) error
func (e *Engine) verifyNodeEventParams(em *eventMethod) error {
	m := em.method.(reflect.Method)

	if m.Type.NumIn() != 4 || m.Type.NumOut() != 1 {
		return fmt.Errorf("warning:事件方法 %s 入参、出参数量不匹配,此函数无法运行", m.Name)
	}

	if m.Type.In(1).Kind().String() != "int" {
		return fmt.Errorf("warning:事件方法 %s 参数1不是int类型,此函数无法运行", m.Name)
	}

	if m.Type.In(2).ConvertibleTo(reflect.TypeFor[*model.Node]()) != true {
		return fmt.Errorf("warning:事件方法 %s 参数2不是*Node类型,此函数无法运行", m.Name)
	}

	if m.Type.In(3).ConvertibleTo(reflect.TypeFor[model.Node]()) != true {
		return fmt.Errorf("warning:事件方法 %s 参数3不是Node类型,此函数无法运行", m.Name)
	}

	if !pkg.TypeIsError(m.Type.Out(0)) {
		return fmt.Errorf("warning:事件方法 %s 返回参数不是error类型,此函数无法运行", m.Name)
	}
	return nil
}

// getEvent 从事件池中获取事件方法。
func (e *Engine) getEvent(name string) (*eventMethod, bool) {
	e.eventPoolMu.RLock()
	defer e.eventPoolMu.RUnlock()
	em, ok := e.eventPool[name]
	return em, ok
}
