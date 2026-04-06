package event

import (
	"context"
	"log/slog"

	"github.com/Bunny3th/easy-workflow/internal/model"
	easyworkflow "github.com/Bunny3th/easy-workflow"
	"gorm.io/gorm"
)

// 这里创建了一个角色-用户的人员库，用来模拟数据库中存储的角色-用户对应关系
var RoleUser = make(map[string][]string)

func init() {
	// 初始化人事数据
	RoleUser["主管"] = []string{"张经理"}
	RoleUser["人事经理"] = []string{"人事老刘"}
	RoleUser["老板"] = []string{"李老板", "老板娘"}
	RoleUser["副总"] = []string{"赵总", "钱总", "孙总"}
}

// 示例事件
type MyEvent struct {
	eng *easyworkflow.Engine
}

// getProcessNameByInstID 根据流程实例ID获取流程名称。
func (e *MyEvent) getProcessNameByInstID(ctx context.Context, instID int) (string, error) {
	var name string
	err := e.eng.DB().WithContext(ctx).Raw("SELECT b.name FROM proc_inst a JOIN proc_def b ON a.proc_id=b.id WHERE a.id=?", instID).Scan(&name).Error
	return name, err
}

// OnEnd 节点结束事件
func (e *MyEvent) OnEnd(ctx context.Context, instID int, currentNode *model.Node, prevNode model.Node) error {
	processName, err := e.getProcessNameByInstID(ctx, instID)
	if err != nil {
		return err
	}
	slog.Info("节点结束", "流程", processName, "节点", currentNode.NodeName)
	return nil
}

// OnNotify 通知
func (e *MyEvent) OnNotify(ctx context.Context, instID int, currentNode *model.Node, prevNode model.Node) error {
	processName, err := e.getProcessNameByInstID(ctx, instID)
	if err != nil {
		return err
	}
	slog.Info("通知节点中对应人员", "流程", processName, "节点", currentNode.NodeName)
	if currentNode.NodeType == model.EndNode {
		slog.Info("流程结束", "流程", processName)
		variables, err := e.eng.ResolveVariables(ctx, model.ResolveVariablesParams{InstanceID: instID, Variables: []string{"$starter"}})
		if err != nil {
			return err
		}
		slog.Info("通知流程创建人流程已完成", "创建人", variables["$starter"], "流程", processName)
	} else {
		for _, user := range currentNode.UserIDs {
			slog.Info("通知用户抓紧去处理", "用户", user)
		}
	}
	return nil
}

// OnResolveRoles 解析角色
func (e *MyEvent) OnResolveRoles(ctx context.Context, instID int, currentNode *model.Node, prevNode model.Node) error {
	processName, err := e.getProcessNameByInstID(ctx, instID)
	if err != nil {
		return err
	}
	slog.Info("开始解析角色", "流程", processName, "节点", currentNode.NodeName)
	for _, role := range currentNode.Roles {
		if users, ok := RoleUser[role]; ok {
			currentNode.UserIDs = append(currentNode.UserIDs, users...)
		}
	}
	return nil
}

// OnTaskForceNodePass 任务事件
// 在示例流程中，"副总审批"是一个会签节点，需要3个副总全部通过，节点才算通过
// 现在通过任务事件改变会签通过人数，设为只要2人通过，即算通过
func (e *MyEvent) OnTaskForceNodePass(ctx context.Context, taskID int, currentNode *model.Node, prevNode model.Node) error {
	taskInfo, err := e.eng.GetTaskInfo(ctx, taskID)
	if err != nil {
		return err
	}

	processName, err := e.getProcessNameByInstID(ctx, taskInfo.ProcInstID)
	if err != nil {
		return err
	}
	slog.Info("任务结束事件", "流程", processName, "节点", currentNode.NodeName, "任务ID", taskInfo.TaskID)

	// 获取节点审批状态：总任务数、通过数、驳回数
	type nodeStatus struct {
		Total    int
		Passed   int
		Rejected int
	}
	var status nodeStatus
	err = e.eng.DB().WithContext(ctx).Raw(
		"SELECT COUNT(*) AS total, SUM(CASE WHEN status=1 THEN 1 ELSE 0 END) AS passed, SUM(CASE WHEN status=2 THEN 1 ELSE 0 END) AS rejected FROM proc_task WHERE proc_inst_id=? AND node_id=? AND batch_code=?",
		taskInfo.ProcInstID, taskInfo.NodeID, taskInfo.BatchCode).Scan(&status).Error
	if err != nil {
		return err
	}

	// 如果通过数>=2，则自动将未通过的任务置为通过
	if status.Passed >= 2 {
		err = e.eng.DB().Transaction(func(tx *gorm.DB) error {
			return tx.Model(&gorm.Model{}).
				Exec("UPDATE proc_task SET comment='通过人数已满2人，系统自动代表你通过', is_finished=1, status=1 WHERE proc_inst_id=? AND node_id=? AND batch_code=? AND is_finished=0",
					taskInfo.ProcInstID, taskInfo.NodeID, taskInfo.BatchCode).Error
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// OnRevoke 撤销事件
func (e *MyEvent) OnRevoke(ctx context.Context, instID int, revokeUserID string) error {
	processName, err := e.getProcessNameByInstID(ctx, instID)
	if err != nil {
		return err
	}
	slog.Info("发起撤销", "流程", processName, "用户", revokeUserID)
	return nil
}

// NewMyEvent 创建并返回示例事件实例。
func NewMyEvent(eng *easyworkflow.Engine) *MyEvent {
	return &MyEvent{eng: eng}
}
