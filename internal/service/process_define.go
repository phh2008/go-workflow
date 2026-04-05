package service

import (
	"context"
	"errors"

	"github.com/Bunny3th/easy-workflow/internal/entity"
	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/pkg"
	"github.com/Bunny3th/easy-workflow/internal/repository"
	"gorm.io/gorm"
)

// ProcessParse 流程定义解析，将 JSON 转换为 Process 结构体。
func (e *Engine) ProcessParse(ctx context.Context, resource string) (*model.Process, error) {
	_ = ctx
	var process model.Process
	err := pkg.JSONToStruct(resource, &process)
	if err != nil {
		return nil, err
	}
	return &process, nil
}

// ProcessSave 流程定义保存，返回流程ID。
func (e *Engine) ProcessSave(ctx context.Context, params model.ProcessSaveParams) (int, error) {
	process, err := e.ProcessParse(ctx, params.Resource)
	if err != nil {
		return 0, err
	}

	if process.ProcessName == "" || process.Source == "" || params.CreateUserID == "" {
		return 0, errors.New("流程名称、来源、创建人ID不能为空")
	}

	// 判断此工作流是否已定义
	procID, version, err := e.repo.GetProcessIDByName(ctx, repository.GetProcIDByNameParams{
		Name:   process.ProcessName,
		Source: process.Source,
	})
	if err != nil {
		return 0, err
	}

	err = e.db.Transaction(func(tx *gorm.DB) error {
		txCtx := repository.WithTx(ctx, tx)

		if procID != 0 {
			// 已有老版本，将老版本移到历史表中
			if err := e.repo.ArchiveProcDef(txCtx, process.ProcessName, process.Source); err != nil {
				return err
			}
			// 更新现有定义
			newVersion := version + 1
			if err := e.repo.UpdateProcDef(txCtx, repository.UpdateProcDefParams{
				Name:     process.ProcessName,
				Source:   process.Source,
				Resource: params.Resource,
				UserID:   params.CreateUserID,
				Version:  newVersion,
			}); err != nil {
				return err
			}
			version = newVersion
		} else {
			// 无老版本，直接插入
			procDef := &entity.ProcDef{
				Name:      process.ProcessName,
				Resource:  params.Resource,
				UserID:    params.CreateUserID,
				Source:    process.Source,
				CreatTime: entity.Now(),
				Version:   1,
			}
			if err := e.repo.CreateProcDef(txCtx, procDef); err != nil {
				return err
			}
			procID = procDef.ID
			version = procDef.Version
		}

		// 将 proc_execution 表对应数据移到历史表
		if err := e.repo.ArchiveExecutions(txCtx, procID); err != nil {
			return err
		}

		// 删除 proc_execution 表对应数据
		if err := e.repo.DeleteExecutions(txCtx, procID); err != nil {
			return err
		}

		// 解析节点之间的关系
		executions := nodesToExecutions(procID, version, process.Nodes)

		// 将执行关系插入 proc_execution 表
		if err := e.repo.SaveExecutions(txCtx, executions); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	// 移除缓存中对应流程ID的内容
	e.procCacheMu.Lock()
	delete(e.procCache, procID)
	e.procCacheMu.Unlock()

	return procID, nil
}

// GetProcessDefine 获取流程定义（通过流程ID）。
func (e *Engine) GetProcessDefine(ctx context.Context, procID int) (*model.Process, error) {
	resource, err := e.repo.GetProcessResource(ctx, procID)
	if err != nil {
		return nil, err
	}
	if resource == "" {
		return nil, errors.New("未找到对应流程定义")
	}
	return e.ProcessParse(ctx, resource)
}

// GetProcessList 获取某个 source 下所有流程定义。
func (e *Engine) GetProcessList(ctx context.Context, source string) ([]entity.ProcDef, error) {
	return e.repo.ListProcessDef(ctx, source)
}

// nodesToExecutions 将 Node 转换为可被数据库表记录的执行步骤。
// 节点的 PrevNodeID 可能是 n 个，则在数据库表中需要存 n 行。
func nodesToExecutions(procID int, procVersion int, nodes []model.Node) []entity.ProcExecution {
	var executions []entity.ProcExecution
	for _, n := range nodes {
		if len(n.PrevNodeIDs) <= 1 {
			var prevNodeID string
			if len(n.PrevNodeIDs) == 0 {
				prevNodeID = ""
			} else {
				prevNodeID = n.PrevNodeIDs[0]
			}
			executions = append(executions, entity.ProcExecution{
				ProcID:      procID,
				ProcVersion: procVersion,
				NodeID:      n.NodeID,
				NodeName:    n.NodeName,
				PrevNodeID:  prevNodeID,
				NodeType:    int(n.NodeType),
				IsCosigned:  n.IsCosigned,
				CreateTime:  entity.Now(),
			})
		} else {
			for _, prev := range n.PrevNodeIDs {
				executions = append(executions, entity.ProcExecution{
					ProcID:      procID,
					ProcVersion: procVersion,
					NodeID:      n.NodeID,
					NodeName:    n.NodeName,
					PrevNodeID:  prev,
					NodeType:    int(n.NodeType),
					IsCosigned:  n.IsCosigned,
					CreateTime:  entity.Now(),
				})
			}
		}
	}
	return executions
}
