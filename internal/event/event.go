package event

import (
	"context"

	"github.com/Bunny3th/easy-workflow/internal/model"
)

// NodeEventHandler 节点事件处理器类型。
// 用于 NodeStartEvents、NodeEndEvents、TaskFinishEvents。
//
// 参数:
//   - ctx: 上下文，支持超时、trace 等
//   - id: 对于 NodeStartEvents/NodeEndEvents 为流程实例 ID，对于 TaskFinishEvents 为任务 ID
//   - currentNode: 当前节点（指针，可修改节点数据如 UserIDs）
//   - prevNode: 上一个节点
type NodeEventHandler func(ctx context.Context, id int, currentNode *model.Node, prevNode model.Node) error

// ProcEventHandler 流程事件处理器类型。
// 用于 RevokeEvents（流程撤销事件）。
//
// 参数:
//   - ctx: 上下文
//   - instID: 流程实例 ID
//   - revokeUserID: 撤销操作人 ID
type ProcEventHandler func(ctx context.Context, instID int, revokeUserID string) error
