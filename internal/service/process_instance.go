package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Bunny3th/easy-workflow/internal/entity"
	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/repository"
	"gorm.io/gorm"
)

// getProcCache 从缓存中获取流程节点定义，缓存未命中则从数据库加载。
func (e *Engine) getProcCache(ctx context.Context, procID int) (map[string]model.Node, error) {
	e.procCacheMu.RLock()
	nodes, ok := e.procCache[procID]
	e.procCacheMu.RUnlock()

	if ok {
		return nodes, nil
	}

	// 缓存未命中，从数据库加载
	process, err := e.GetProcessDefine(ctx, procID)
	if err != nil {
		return nil, err
	}

	pn := make(map[string]model.Node)
	for _, n := range process.Nodes {
		pn[n.NodeID] = n
	}

	e.procCacheMu.Lock()
	e.procCache[procID] = pn
	e.procCacheMu.Unlock()

	return pn, nil
}

// instanceInit 流程实例初始化：创建实例记录、保存变量、获取开始节点。
// 返回流程实例ID和开始节点。
func (e *Engine) instanceInit(ctx context.Context, procID int, businessID string, variablesJSON string) (int, *model.Node, error) {
	// 解析变量
	vars, err := e.ParseVariable(ctx, variablesJSON)
	if err != nil {
		return 0, nil, err
	}
	varsMap := make(map[string]string)
	for _, v := range vars {
		varsMap[v.Key] = v.Value
	}
	// 获取流程定义（流程中所有node）
	nodes, err := e.getProcCache(ctx, procID)
	if err != nil {
		return 0, nil, err
	}

	// 检查流程节点中的事件是否都已经注册
	if err := e.verifyEvents(ctx, procID, nodes); err != nil {
		return 0, nil, err
	}

	// 获取流程开始节点ID
	startNodeID, err := e.repo.GetStartNodeID(ctx, procID)
	if err != nil {
		return 0, nil, err
	}
	if startNodeID == "" {
		return 0, nil, fmt.Errorf("无法获取流程ID为%d的开始节点", procID)
	}

	// 获得开始节点
	startNode := nodes[startNodeID]

	// 获取流程定义版本号
	var procDef entity.ProcDef
	if err := e.db.WithContext(ctx).Where("id=?", procID).First(&procDef).Error; err != nil {
		return 0, nil, err
	}

	var instID int

	// 开启事务
	err = e.db.Transaction(func(tx *gorm.DB) error {
		txCtx := repository.WithTx(ctx, tx)

		// 在实例表中生成一条数据
		procInst := &entity.ProcInst{
			ProcID:        procID,
			ProcVersion:   procDef.Version,
			BusinessID:    businessID,
			CurrentNodeID: startNode.NodeID,
		}
		if err := e.repo.CreateInstance(txCtx, procInst); err != nil {
			return err
		}

		instID = procInst.ID

		// 保存流程变量
		if err := e.repo.SaveVariable(txCtx, procInst.ID, vars); err != nil {
			return err
		}

		// 获取流程起始人
		users, err := e.resolveNodeUsersFromVars(txCtx, startNode, varsMap)
		if err != nil {
			return err
		}

		if len(users) < 1 {
			return errors.New("未指定处理人，无法处理节点:" + startNode.NodeName)
		}

		// 更新起始人到流程实例表
		if err := e.repo.UpdateInstance(txCtx, procInst.ID, map[string]any{"starter": users[0]}); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, &startNode, err
	}

	return instID, &startNode, nil
}

// InstanceStart 启动流程实例，返回流程实例ID。
func (e *Engine) InstanceStart(ctx context.Context, req model.InstanceStartReq) (int, error) {
	instID, startNode, err := e.instanceInit(ctx, req.ProcessID, req.BusinessID, req.VariablesJSON)
	if err != nil {
		return 0, err
	}

	err = e.startNodeHandle(ctx, instID, startNode, req.Comment, req.VariablesJSON)
	if err != nil {
		// 删除已建立的实例记录、变量记录、任务记录
		e.db.WithContext(ctx).Where("id=?", instID).Delete(&entity.ProcInst{})
		e.db.WithContext(ctx).Where("proc_inst_id=?", instID).Delete(&entity.ProcInstVariable{})
		e.db.WithContext(ctx).Where("proc_inst_id=?", instID).Delete(&entity.ProcTask{})
		return instID, err
	}

	return instID, nil
}

// InstanceRevoke 撤销流程实例。
// req.Force 为 false 时，只有流程回到发起人这里才能撤销。
func (e *Engine) InstanceRevoke(ctx context.Context, req model.InstanceRevokeReq) error {
	if !req.Force {
		// 判断当前Node是否是开始Node
		sql := "SELECT a.id FROM proc_inst a " +
			"JOIN proc_execution b ON a.proc_id=b.proc_id AND a.current_node_id=b.node_id " +
			"WHERE a.id=? AND (b.prev_node_id IS NULL OR b.prev_node_id='') LIMIT 1"
		var id int
		if err := e.db.WithContext(ctx).Raw(sql, req.InstanceID).Scan(&id).Error; err != nil {
			return err
		}
		if id == 0 {
			return errors.New("当前流程所在节点不是发起节点，无法撤销!")
		}
	}

	// 获取流程ID
	procID, err := e.repo.GetProcessIDByInstID(ctx, req.InstanceID)
	if err != nil {
		return err
	}

	// 获取流程定义
	process, err := e.GetProcessDefine(ctx, procID)
	if err != nil {
		return err
	}

	// 执行流程撤销事件
	if err := e.runProcEvents(ctx, process.RevokeEvents, req.InstanceID, req.RevokeUserID); err != nil {
		return err
	}

	// 调用 endNodeHandle，做数据清理归档
	return e.endNodeHandle(ctx, req.InstanceID, 2)
}

// GetInstanceInfo 获取流程实例信息。
func (e *Engine) GetInstanceInfo(ctx context.Context, instID int) (model.InstanceView, error) {
	return e.repo.GetInstanceInfo(ctx, instID)
}

// GetInstanceStartByUser 获取起始人为特定用户的流程实例（含分页和总数）。
func (e *Engine) GetInstanceStartByUser(ctx context.Context, req model.InstanceListReq) (*model.PageData[model.InstanceView], error) {
	instances, err := e.repo.ListInstanceStartByUser(ctx, repository.ListInstByUserParams{
		UserID:      req.UserID,
		ProcessName: req.ProcessName,
		Offset:      req.Offset(),
		Limit:       req.GetPageSize(),
	})
	if err != nil {
		return nil, err
	}
	count, err := e.repo.CountInstanceStartByUser(ctx, repository.CountByUserParams{
		UserID:      req.UserID,
		ProcessName: req.ProcessName,
	})
	if err != nil {
		return nil, err
	}
	return model.NewPageData[model.InstanceView](req.GetPageNo(), req.GetPageSize()).SetData(instances).SetCount(count), nil
}

// startNodeHandle 开始节点处理。开始节点是一个特殊的任务节点：
// 1. 在生成流程实例的同时，就要运行开始节点
// 2. 开始节点生成的任务自动完成，而后自动进行下一个节点的处理
func (e *Engine) startNodeHandle(ctx context.Context, instID int, startNode *model.Node, comment, variablesJSON string) error {
	slog.Debug("[startNodeHandle] 开始处理开始节点", "instID", instID, "nodeID", startNode.NodeID, "nodeName", startNode.NodeName)

	if startNode.NodeType != model.RootNode {
		return errors.New("不是开始节点，无法处理节点:" + startNode.NodeName)
	}

	// 处理节点开始事件
	if err := e.runNodeEvents(ctx, startNode.NodeStartEvents, instID, startNode, model.Node{}); err != nil {
		return err
	}

	// 生成Task
	taskIDs, err := e.taskNodeHandle(ctx, instID, startNode, model.Node{})
	if err != nil {
		slog.Debug("[startNodeHandle] taskNodeHandle 失败", "error", err)
		return err
	}
	slog.Debug("[startNodeHandle] 开始节点任务已生成", "instID", instID, "taskIDs", taskIDs)

	// 完成task，并获取下一步NodeID
	if err := e.TaskPass(ctx, model.TaskActionReq{TaskID: taskIDs[0], Comment: comment, VariableJSON: variablesJSON}, false); err != nil {
		slog.Debug("[startNodeHandle] TaskPass 失败", "error", err)
		return err
	}
	slog.Debug("[startNodeHandle] 开始节点任务已自动通过", "instID", instID)

	return nil
}
