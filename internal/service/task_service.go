package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Bunny3th/easy-workflow/internal/entity"
	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/pkg"
	"github.com/Bunny3th/easy-workflow/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// taskOption 任务处理选项。
type taskOption struct {
	status                  int  // 任务状态 1:通过 2:驳回
	directlyToWhoRejectedMe bool // 任务通过(pass)时直接返回到上一个驳回我的节点
}

// TaskPass 通过任务，在本节点处理完毕的情况下会自动处理下一个节点。
// directlyToRejected 为 true 时，直接跳到上一个驳回自己的节点。
func (e *Engine) TaskPass(ctx context.Context, req model.TaskActionReq, directlyToRejected bool) error {
	return e.processTask(ctx, req.TaskID, req.Comment, req.VariableJSON, taskOption{status: 1, directlyToWhoRejectedMe: directlyToRejected})
}

// TaskReject 驳回任务，在本节点处理完毕的情况下会自动处理下一个节点。
func (e *Engine) TaskReject(ctx context.Context, req model.TaskActionReq) error {
	return e.processTask(ctx, req.TaskID, req.Comment, req.VariableJSON, taskOption{status: 2})
}

// TaskTransfer 将任务转交给其他用户处理。
func (e *Engine) TaskTransfer(ctx context.Context, req model.TaskTransferReq) error {
	users := pkg.MakeUnique(req.Users)
	if len(users) < 1 {
		return errors.New("转让任务操作必须指定至少一个候选人")
	}

	taskInfo, err := e.GetTaskInfo(ctx, req.TaskID)
	if err != nil {
		return err
	}

	if taskInfo.IsFinished == 1 {
		return errors.New("任务已完成，无法转交")
	}

	err = e.db.Transaction(func(tx *gorm.DB) error {
		txCtx := repository.WithTx(ctx, tx)

		// 删除原任务
		if err := e.repo.DeleteTaskByID(txCtx, req.TaskID); err != nil {
			return err
		}

		// 生成新任务
		var tasks []entity.ProcTask
		for _, u := range users {
			tasks = append(tasks, entity.ProcTask{
				ProcID:             taskInfo.ProcID,
				ProcInstID:         taskInfo.ProcInstID,
				ProcInstCreateTime: *taskInfo.ProcInstCreateTime,
				BusinessID:         taskInfo.BusinessID,
				Starter:            taskInfo.Starter,
				NodeID:             taskInfo.NodeID,
				NodeName:           taskInfo.NodeName,
				PrevNodeID:         taskInfo.PrevNodeID,
				IsCosigned:         taskInfo.IsCosigned,
				BatchCode:          taskInfo.BatchCode,
				UserID:             u,
			})
		}

		if err := e.repo.CreateTasks(txCtx, tasks); err != nil {
			return err
		}

		return nil
	})

	return err
}

// TaskFreeReject 自由驳回到任意一个上游节点。
func (e *Engine) TaskFreeReject(ctx context.Context, req model.TaskFreeRejectReq) error {
	taskInfo, err := e.GetTaskInfo(ctx, req.TaskID)
	if err != nil {
		return err
	}

	if taskInfo.IsFinished == 1 {
		return fmt.Errorf("节点ID%d已处理，无需操作", req.TaskID)
	}

	currentNode, err := e.getInstanceNode(ctx, taskInfo.ProcInstID, taskInfo.NodeID)
	if err != nil {
		return err
	}

	if currentNode.NodeType == model.RootNode {
		return errors.New("起始节点无法驳回")
	}

	rejectToNode, err := e.getInstanceNode(ctx, taskInfo.ProcInstID, req.RejectToNodeID)
	if err != nil {
		return err
	}

	if err := e.taskSubmit(ctx, taskInfo, req.Comment, req.VariableJSON, 2); err != nil {
		return err
	}

	if err := e.processNode(ctx, taskInfo.ProcInstID, &rejectToNode, currentNode); err != nil {
		if revokeErr := e.taskRevoke(ctx, req.TaskID); revokeErr != nil {
			slog.Error("[TaskFreeReject] taskRevoke 失败", "taskID", req.TaskID, "error", revokeErr)
		}
		return err
	}

	return nil
}

// GetTaskInfo 获取任务信息。
func (e *Engine) GetTaskInfo(ctx context.Context, taskID int) (model.TaskView, error) {
	return e.repo.GetTaskInfo(ctx, taskID)
}

// GetTaskToDoList 获取特定用户待办任务列表（含分页和总数）。
func (e *Engine) GetTaskToDoList(ctx context.Context, req model.TaskListReq) (*model.PageData[model.TaskView], error) {
	tasks, err := e.repo.ListTaskToDo(ctx, repository.ListToDoParams{
		UserID:      req.UserID,
		ProcessName: req.ProcessName,
		Asc:         req.Asc,
		Offset:      req.Offset(),
		Limit:       req.GetPageSize(),
	})
	if err != nil {
		return nil, err
	}
	count, err := e.repo.CountTaskToDo(ctx, repository.CountByUserParams{
		UserID:      req.UserID,
		ProcessName: req.ProcessName,
	})
	if err != nil {
		return nil, err
	}
	return model.NewPageData[model.TaskView](req.GetPageNo(), req.GetPageSize()).SetData(tasks).SetCount(count), nil
}

// GetTaskFinishedList 获取特定用户已完成任务列表（含分页和总数）。
func (e *Engine) GetTaskFinishedList(ctx context.Context, req model.TaskFinishedListReq) (*model.PageData[model.TaskView], error) {
	tasks, err := e.repo.ListTaskFinished(ctx, repository.ListFinishedParams{
		UserID:          req.UserID,
		ProcessName:     req.ProcessName,
		IgnoreStartByMe: req.IgnoreStartByMe,
		Asc:             req.Asc,
		Offset:          req.Offset(),
		Limit:           req.GetPageSize(),
	})
	if err != nil {
		return nil, err
	}
	count, err := e.repo.CountTaskFinished(ctx, repository.CountFinishedParams{
		UserID:          req.UserID,
		ProcessName:     req.ProcessName,
		IgnoreStartByMe: req.IgnoreStartByMe,
	})
	if err != nil {
		return nil, err
	}
	return model.NewPageData[model.TaskView](req.GetPageNo(), req.GetPageSize()).SetData(tasks).SetCount(count), nil
}

// TaskUpstreamNodeList 根据流程定义，列出 task 所在节点的所有上游节点。
func (e *Engine) TaskUpstreamNodeList(ctx context.Context, taskID int) ([]model.Node, error) {
	task, err := e.GetTaskInfo(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return e.repo.GetUpstreamNodes(ctx, task.NodeID)
}

// GetInstanceTaskHistory 获取流程实例下任务历史记录。
func (e *Engine) GetInstanceTaskHistory(ctx context.Context, instID int) ([]model.TaskView, error) {
	return e.repo.ListInstanceTaskHistory(ctx, instID)
}

// WhatCanIDo 判断某个任务可以执行哪些操作。
func (e *Engine) WhatCanIDo(ctx context.Context, taskID int) (model.TaskAction, error) {
	act := model.TaskAction{
		CanPass:                     true,
		CanReject:                   true,
		CanFreeRejectToUpstreamNode: true,
		CanDirectlyToWhoRejectedMe:  true,
		CanRevoke:                   false,
	}

	taskInfo, err := e.GetTaskInfo(ctx, taskID)
	if err != nil {
		return model.TaskAction{}, err
	}

	if taskInfo.IsFinished == 1 {
		return model.TaskAction{}, nil
	}

	node, err := e.getInstanceNode(ctx, taskInfo.ProcInstID, taskInfo.NodeID)
	if err != nil {
		return model.TaskAction{}, err
	}

	// 起始节点不能做驳回动作
	if node.NodeType == model.RootNode {
		act.CanReject = false
		act.CanFreeRejectToUpstreamNode = false
		act.CanRevoke = true
	}

	// 会签节点不能使用 DirectlyToWhoRejectedMe
	if taskInfo.IsCosigned == 1 {
		act.CanDirectlyToWhoRejectedMe = false
	}

	// 上一个节点未做驳回
	prevReject, err := e.isPrevNodeRejected(ctx, taskInfo)
	if err != nil {
		return model.TaskAction{}, err
	}
	if !prevReject {
		act.CanDirectlyToWhoRejectedMe = false
	}

	return act, nil
}

// createTask 生成任务，返回生成的任务ID数组。
func (e *Engine) createTask(ctx context.Context, instID int, nodeID string, prevNodeID string, userIDs []string) ([]int, error) {
	// 获取本节点中未结束任务的用户
	notFinishUsers, err := e.repo.GetNotFinishUsers(ctx, repository.NodeQueryParams{InstID: instID, NodeID: nodeID})
	if err != nil {
		return nil, err
	}

	// 先去重
	userIDs = pkg.MakeUnique(userIDs)

	// 去掉本节点中未结束任务的用户
	for _, nu := range notFinishUsers {
		userIDs = pkg.RemoveAllElements(userIDs, nu)
	}

	// 获取流程ID
	procID, err := e.repo.GetProcessIDByInstID(ctx, instID)
	if err != nil {
		return nil, err
	}

	// 获取Node信息
	node, err := e.getInstanceNode(ctx, instID, nodeID)
	if err != nil {
		return nil, err
	}

	// 获取实例创建时间和业务ID
	var procInst entity.ProcInst
	if err := e.db.WithContext(ctx).Where("id=?", instID).First(&procInst).Error; err != nil {
		return nil, err
	}

	// 生成批次码
	batchCode, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	// 生成任务数据
	var tasks []entity.ProcTask
	for _, u := range userIDs {
		tasks = append(tasks, entity.ProcTask{
			ProcID:             procID,
			ProcInstID:         instID,
			ProcInstCreateTime: procInst.CreateTime,
			BusinessID:         procInst.BusinessID,
			Starter:            procInst.Starter,
			NodeID:             nodeID,
			NodeName:           node.NodeName,
			PrevNodeID:         prevNodeID,
			IsCosigned:         node.IsCosigned,
			BatchCode:          batchCode.String(),
			UserID:             u,
		})
	}

	var taskIDs []int

	err = e.db.Transaction(func(tx *gorm.DB) error {
		txCtx := repository.WithTx(ctx, tx)

		if err := e.repo.CreateTasks(txCtx, tasks); err != nil {
			return err
		}

		// 更新 proc_inst 表 current_node_id 字段
		if err := e.repo.UpdateInstance(txCtx, instID, map[string]any{"current_node_id": nodeID}); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	for _, t := range tasks {
		taskIDs = append(taskIDs, t.ID)
	}

	return taskIDs, nil
}

// processTask 任务处理（通过/驳回）。
func (e *Engine) processTask(ctx context.Context, taskID int, comment, varJSON string, option taskOption) error {
	taskInfo, err := e.GetTaskInfo(ctx, taskID)
	if err != nil {
		return err
	}

	currentNode, err := e.getInstanceNode(ctx, taskInfo.ProcInstID, taskInfo.NodeID)
	if err != nil {
		return err
	}

	slog.Debug("[processTask] 开始处理任务", "taskID", taskID, "instID", taskInfo.ProcInstID, "nodeID", taskInfo.NodeID, "nodeName", taskInfo.NodeName, "userID", taskInfo.UserID, "isFinished", taskInfo.IsFinished, "status", option.status)

	if taskInfo.IsFinished == 1 {
		return fmt.Errorf("节点ID%d已处理，无需操作", taskID)
	}

	// 如果是通过且开启 DirectlyToWhoRejectedMe，做功能前置验证
	if option.status == 1 && option.directlyToWhoRejectedMe {
		if taskInfo.IsCosigned == 1 {
			return errors.New("会签节点无法使用【DirectlyToWhoRejectedMe】功能")
		}
		if taskInfo.PrevNodeID == "" {
			return errors.New("此任务不存在上级节点,无法使用【DirectlyToWhoRejectedMe】功能")
		}
		prevReject, err := e.isPrevNodeRejected(ctx, taskInfo)
		if err != nil {
			return err
		}
		if !prevReject {
			return errors.New("此任务的上一节点并未做驳回,无法使用【DirectlyToWhoRejectedMe】功能")
		}
	}

	// 驳回时，起始节点不能做驳回
	if option.status == 2 && currentNode.NodeType == model.RootNode {
		return errors.New("起始节点无法驳回")
	}

	if err := e.taskSubmit(ctx, taskInfo, comment, varJSON, option.status); err != nil {
		return err
	}

	// 同步更新 taskInfo 状态，因为 taskSubmit 更新了数据库但 taskInfo 是值拷贝
	taskInfo.Status = option.status
	taskInfo.IsFinished = 1

	// 当前 task 上一个节点
	var prevNode model.Node
	if taskInfo.PrevNodeID == "" {
		prevNode = model.Node{}
	} else {
		prevNode, err = e.getInstanceNode(ctx, taskInfo.ProcInstID, taskInfo.PrevNodeID)
		if err != nil {
			if revokeErr := e.taskRevoke(ctx, taskID); revokeErr != nil {
				slog.Error("[processTask] taskRevoke 失败", "taskID", taskID, "error", revokeErr)
			}
			return err
		}
	}

	// 处理任务结束事件
	if err := e.runNodeEvents(ctx, currentNode.TaskFinishEvents, taskInfo.TaskID, &currentNode, prevNode); err != nil {
		if revokeErr := e.taskRevoke(ctx, taskID); revokeErr != nil {
			slog.Error("[processTask] taskRevoke 失败", "taskID", taskID, "error", revokeErr)
		}
		return err
	}

	// 获取任务执行完毕后下一个节点
	var nextNode model.Node
	if option.status == 1 && option.directlyToWhoRejectedMe {
		nextNode, err = e.getInstanceNode(ctx, taskInfo.ProcInstID, taskInfo.PrevNodeID)
		if err != nil {
			if revokeErr := e.taskRevoke(ctx, taskID); revokeErr != nil {
				slog.Error("[processTask] taskRevoke 失败", "taskID", taskID, "error", revokeErr)
			}
			return err
		}
	} else {
		nextNode, err = e.taskNextNode(ctx, taskInfo)
		if err != nil {
			slog.Debug("[processTask] taskNextNode 失败", "error", err)
			if revokeErr := e.taskRevoke(ctx, taskID); revokeErr != nil {
				slog.Error("[processTask] taskRevoke 失败", "taskID", taskID, "error", revokeErr)
			}
			return err
		}
	}

	slog.Debug("[processTask] 下一个节点", "taskID", taskID, "nextNodeID", nextNode.NodeID, "nextNodeName", nextNode.NodeName, "nextNodeType", nextNode.NodeType)

	if nextNode.NodeID == "" {
		slog.Debug("[processTask] 无下一节点，流程在此终止", "taskID", taskID)
		return nil
	}

	// 处理节点结束事件
	if err := e.runNodeEvents(ctx, currentNode.NodeEndEvents, taskInfo.ProcInstID, &currentNode, prevNode); err != nil {
		if revokeErr := e.taskRevoke(ctx, taskID); revokeErr != nil {
			slog.Error("[processTask] taskRevoke 失败", "taskID", taskID, "error", revokeErr)
		}
		return err
	}

	// 开始处理下一个节点
	if err := e.processNode(ctx, taskInfo.ProcInstID, &nextNode, currentNode); err != nil {
		slog.Debug("[processTask] processNode 失败", "nextNodeID", nextNode.NodeID, "error", err)
		if revokeErr := e.taskRevoke(ctx, taskID); revokeErr != nil {
			slog.Error("[processTask] taskRevoke 失败", "taskID", taskID, "error", revokeErr)
		}
		return err
	}

	return nil
}

// taskSubmit 将任务提交数据保存到数据库。
func (e *Engine) taskSubmit(ctx context.Context, taskInfo model.TaskView, comment, varJSON string, status int) error {
	if taskInfo.IsFinished == 1 {
		return fmt.Errorf("节点ID%d已处理，无需操作", taskInfo.TaskID)
	}
	// 解析变量
	vars, err := e.ParseVariable(ctx, varJSON)
	if err != nil {
		return err
	}
	return e.db.Transaction(func(tx *gorm.DB) error {
		txCtx := repository.WithTx(ctx, tx)
		// 更新 task 表记录
		if err := e.repo.UpdateTask(txCtx, taskInfo.TaskID, map[string]any{
			"status":        status,
			"is_finished":   1,
			"comment":       comment,
			"finished_time": entity.Now(),
		}); err != nil {
			return err
		}
		// 非会签通过 或 任何驳回：将同一批次 task 的 is_finished 设为 1
		if (taskInfo.IsCosigned == 0 && status == 1) || status == 2 {
			if err := e.repo.UpdateTasksByBatchCode(txCtx, taskInfo.BatchCode, map[string]any{
				"is_finished":   1,
				"finished_time": entity.Now(),
			}); err != nil {
				return err
			}
		}
		// 设置实例变量
		if err := e.repo.SaveVariable(txCtx, taskInfo.ProcInstID, vars); err != nil {
			return err
		}
		return nil
	})
}

// taskRevoke 回滚已提交的任务状态。
func (e *Engine) taskRevoke(ctx context.Context, taskID int) error {
	return e.db.Transaction(func(tx *gorm.DB) error {
		txCtx := repository.WithTx(ctx, tx)

		if err := e.repo.RevokeTask(txCtx, taskID); err != nil {
			return err
		}

		// 重置同批次中未做通过或驳回的 task
		if err := tx.Exec("UPDATE proc_task SET is_finished=0 "+
			"WHERE `status`=0 AND batch_code IN (SELECT batch_code FROM proc_task WHERE id=?)", taskID).Error; err != nil {
			return err
		}

		return nil
	})
}

// isPrevNodeRejected 判断任务的上一个节点是否做了驳回。
func (e *Engine) isPrevNodeRejected(ctx context.Context, taskInfo model.TaskView) (bool, error) {
	batchCode, err := e.repo.GetPrevNodeBatchCode(ctx, taskInfo.TaskID)
	if err != nil {
		return false, err
	}

	hasReject, err := e.repo.HasRejectInBatch(ctx, batchCode)
	if err != nil {
		return false, err
	}

	return hasReject, nil
}

// taskNextNode 获取任务执行完毕后下一个节点。
func (e *Engine) taskNextNode(ctx context.Context, taskInfo model.TaskView) (model.Node, error) {
	totalTask, totalPassed, _, err := e.getTaskNodeStatus(ctx, taskInfo)
	if err != nil {
		return model.Node{}, err
	}

	slog.Debug("[taskNextNode] 获取下一节点", "taskID", taskInfo.TaskID, "nodeID", taskInfo.NodeID, "status", taskInfo.Status, "isCosigned", taskInfo.IsCosigned, "totalTask", totalTask, "totalPassed", totalPassed)

	if taskInfo.Status == 1 {
		// 通过的情况
		if taskInfo.IsCosigned == 0 || (taskInfo.IsCosigned == 1 && totalTask == totalPassed) {
			nextNodeID, err := e.repo.GetNextNodeIDByPrevNodeID(ctx, taskInfo.NodeID)
			if err != nil {
				slog.Debug("[taskNextNode] GetNextNodeIDByPrevNodeID 失败", "prevNodeID", taskInfo.NodeID, "error", err)
				return model.Node{}, err
			}
			slog.Debug("[taskNextNode] 查到下一节点", "prevNodeID", taskInfo.NodeID, "nextNodeID", nextNodeID)
			return e.getInstanceNode(ctx, taskInfo.ProcInstID, nextNodeID)
		}
		slog.Debug("[taskNextNode] 会签节点未全部通过，无下一节点", "taskID", taskInfo.TaskID)
		return model.Node{}, nil
	}

	if taskInfo.Status == 2 {
		// 驳回的情况
		var prevNodeID string

		prevReject, err := e.isPrevNodeRejected(ctx, taskInfo)
		if err != nil {
			return model.Node{}, err
		}

		if !prevReject {
			prevNodeID = taskInfo.PrevNodeID
		} else {
			nodes, err := e.repo.GetUpstreamNodes(ctx, taskInfo.NodeID)
			if err != nil {
				return model.Node{}, err
			}
			if len(nodes) == 0 {
				return model.Node{}, errors.New("无法获取上游节点")
			}
			prevNodeID = nodes[0].NodeID
		}

		return e.getInstanceNode(ctx, taskInfo.ProcInstID, prevNodeID)
	}

	return model.Node{}, nil
}

// getTaskNodeStatus 获取任务节点审批状态，返回总任务数、通过数、驳回数。
func (e *Engine) getTaskNodeStatus(ctx context.Context, taskInfo model.TaskView) (int, int, int, error) {
	return e.repo.GetTaskNodeStatus(ctx, repository.TaskNodeStatusParams{InstID: taskInfo.ProcInstID, NodeID: taskInfo.NodeID, BatchCode: taskInfo.BatchCode})
}
