package schedule

import (
	"context"

	"github.com/Bunny3th/easy-workflow"
	"github.com/Bunny3th/easy-workflow/internal/model"
)

// AutoFinishTask 定义一个任务计划：对于UserID为"-1"的任务做自动通过。
// 返回一个 func() error 闭包，用于传入 eng.ScheduleTask()。
func AutoFinishTask(eng *easyworkflow.Engine) func() error {
	return func() error {
		ctx := context.Background()
		// 首先获取所有用户ID为"-1"，且还未完成的任务
		var tasks []model.TaskView
		err := eng.DB().WithContext(ctx).Raw("SELECT * FROM proc_task WHERE user_id='-1' AND is_finished=0").Scan(&tasks).Error
		if err != nil {
			return err
		}

		for _, task := range tasks {
			type node struct {
				NodeID string `gorm:"column:node_id"`
			}
			var PrevNodes []node
			var PrevNodeIDs = make(map[string]any)
			err := eng.DB().WithContext(ctx).Raw(
				"SELECT prev_node_id AS node_id FROM proc_execution WHERE proc_id=? AND node_id=?",
				task.ProcID, task.NodeID).Scan(&PrevNodes).Error
			if err != nil {
				return err
			}
			for _, n := range PrevNodes {
				PrevNodeIDs[n.NodeID] = nil
			}

			if _, ok := PrevNodeIDs[task.PrevNodeID]; ok {
				err = eng.TaskPass(ctx, model.TaskPassParams{TaskID: task.TaskID, Comment: "免审自动通过"})
			} else {
				err = eng.TaskReject(ctx, model.TaskRejectParams{TaskID: task.TaskID, Comment: "免审自动驳回"})
			}
			if err != nil {
				return err
			}
		}
		return nil
	}
}
