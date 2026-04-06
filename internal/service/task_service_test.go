package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/repository"
)

// ---- GetTaskInfo ----

func TestGetTaskInfo(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	expected := model.TaskView{
		TaskID:     10,
		NodeID:     "Approve",
		NodeName:   "审批",
		UserID:     "user1",
		IsFinished: 0,
		Status:     0,
		BatchCode:  "batch-001",
		ProcInstID: 1,
		PrevNodeID: "Start",
		IsCosigned: 0,
	}
	repo.GetTaskInfoFunc = func(ctx context.Context, taskID int) (model.TaskView, error) {
		if taskID != 10 {
			t.Errorf("taskID = %d, 期望 10", taskID)
		}
		return expected, nil
	}

	result, err := eng.GetTaskInfo(ctx, 10)
	if err != nil {
		t.Fatalf("GetTaskInfo 返回错误: %v", err)
	}
	if result.NodeID != "Approve" {
		t.Errorf("NodeID = %q, 期望 %q", result.NodeID, "Approve")
	}
	if result.UserID != "user1" {
		t.Errorf("UserID = %q, 期望 %q", result.UserID, "user1")
	}
}

// ---- GetTaskToDoList ----

func TestGetTaskToDoList(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	expected := []model.TaskView{
		{TaskID: 1, NodeName: "审批1", UserID: "user1"},
		{TaskID: 2, NodeName: "审批2", UserID: "user1"},
	}
	callCount := 0
	repo.ListTaskToDoFunc = func(ctx context.Context, p repository.ListToDoParams) ([]model.TaskView, error) {
		callCount++
		if p.UserID != "user1" {
			t.Errorf("userID = %q, 期望 %q", p.UserID, "user1")
		}
		if p.ProcessName != "请假" {
			t.Errorf("processName = %q, 期望 %q", p.ProcessName, "请假")
		}
		if p.Asc != true {
			t.Errorf("asc = %v, 期望 true", p.Asc)
		}
		return expected, nil
	}
	repo.CountTaskToDoFunc = func(ctx context.Context, p repository.CountByUserParams) (int64, error) {
		return 2, nil
	}

	result, err := eng.GetTaskToDoList(ctx, model.TaskListReq{PageQuery: model.PageQuery{PageNo: 1, PageSize: 10}, UserID: "user1", ProcessName: "请假", Asc: true})
	if err != nil {
		t.Fatalf("GetTaskToDoList 返回错误: %v", err)
	}
	if len(result.Data) != 2 {
		t.Fatalf("返回数量 = %d, 期望 2", len(result.Data))
	}
	if result.Count != 2 {
		t.Errorf("Count = %d, 期望 2", result.Count)
	}
	if callCount != 1 {
		t.Errorf("ListTaskToDo 调用次数 = %d, 期望 1", callCount)
	}
}

// ---- GetTaskFinishedList ----

func TestGetTaskFinishedList(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	expected := []model.TaskView{
		{TaskID: 3, NodeName: "已审批1", UserID: "user2"},
	}
	repo.ListTaskFinishedFunc = func(ctx context.Context, p repository.ListFinishedParams) ([]model.TaskView, error) {
		if p.UserID != "user2" {
			t.Errorf("userID = %q, 期望 %q", p.UserID, "user2")
		}
		if p.IgnoreStartByMe != true {
			t.Errorf("ignoreStartByMe = %v, 期望 true", p.IgnoreStartByMe)
		}
		return expected, nil
	}
	repo.CountTaskFinishedFunc = func(ctx context.Context, p repository.CountFinishedParams) (int64, error) {
		return 1, nil
	}

	result, err := eng.GetTaskFinishedList(ctx, model.TaskFinishedListReq{TaskListReq: model.TaskListReq{PageQuery: model.PageQuery{PageNo: 1, PageSize: 20}, UserID: "user2", ProcessName: "", Asc: false}, IgnoreStartByMe: true})
	if err != nil {
		t.Fatalf("GetTaskFinishedList 返回错误: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("返回数量 = %d, 期望 1", len(result.Data))
	}
	if result.Count != 1 {
		t.Errorf("Count = %d, 期望 1", result.Count)
	}
}

// ---- TaskUpstreamNodeList ----

func TestTaskUpstreamNodeList(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetTaskInfoFunc = func(ctx context.Context, taskID int) (model.TaskView, error) {
		return model.TaskView{TaskID: 5, NodeID: "Approve"}, nil
	}
	expectedNodes := []model.Node{
		{NodeID: "Start", NodeName: "开始", NodeType: model.RootNode},
		{NodeID: "GW1", NodeName: "网关1", NodeType: model.GateWayNode},
	}
	repo.GetUpstreamNodesFunc = func(ctx context.Context, nodeID string) ([]model.Node, error) {
		if nodeID != "Approve" {
			t.Errorf("nodeID = %q, 期望 %q", nodeID, "Approve")
		}
		return expectedNodes, nil
	}

	nodes, err := eng.TaskUpstreamNodeList(ctx, 5)
	if err != nil {
		t.Fatalf("TaskUpstreamNodeList 返回错误: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("返回数量 = %d, 期望 2", len(nodes))
	}
	if nodes[0].NodeID != "Start" {
		t.Errorf("nodes[0].NodeID = %q, 期望 %q", nodes[0].NodeID, "Start")
	}
}

func TestTaskUpstreamNodeList_TaskInfoError(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetTaskInfoFunc = func(ctx context.Context, taskID int) (model.TaskView, error) {
		return model.TaskView{}, errors.New("任务不存在")
	}

	_, err := eng.TaskUpstreamNodeList(ctx, 999)
	if err == nil {
		t.Fatal("TaskUpstreamNodeList 任务不存在时应返回错误")
	}
}

// ---- GetInstanceTaskHistory ----

func TestGetInstanceTaskHistory(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	expected := []model.TaskView{
		{TaskID: 1, NodeName: "开始", Status: 1},
		{TaskID: 2, NodeName: "审批", Status: 1},
	}
	repo.ListInstanceTaskHistoryFunc = func(ctx context.Context, instID int) ([]model.TaskView, error) {
		if instID != 100 {
			t.Errorf("instID = %d, 期望 100", instID)
		}
		return expected, nil
	}

	result, err := eng.GetInstanceTaskHistory(ctx, 100)
	if err != nil {
		t.Fatalf("GetInstanceTaskHistory 返回错误: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("返回数量 = %d, 期望 2", len(result))
	}
}

// ---- WhatCanIDo ----

func TestWhatCanIDo_RootNode(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetTaskInfoFunc = func(ctx context.Context, taskID int) (model.TaskView, error) {
		return model.TaskView{
			TaskID:     1,
			NodeID:     "Start",
			IsFinished: 0,
			ProcInstID: 1,
			IsCosigned: 0,
		}, nil
	}

	// getInstanceNode 需要 GetProcessIDByInstID + procCache
	repo.GetProcessIDByInstIDFunc = func(ctx context.Context, instID int) (int, error) {
		return 1, nil
	}
	eng.procCache[1] = map[string]model.Node{
		"Start": {NodeID: "Start", NodeType: model.RootNode},
	}

	act, err := eng.WhatCanIDo(ctx, 1)
	if err != nil {
		t.Fatalf("WhatCanIDo 返回错误: %v", err)
	}
	if act.CanReject {
		t.Error("起始节点 CanReject 应为 false")
	}
	if act.CanFreeRejectToUpstreamNode {
		t.Error("起始节点 CanFreeRejectToUpstreamNode 应为 false")
	}
	if !act.CanRevoke {
		t.Error("起始节点 CanRevoke 应为 true")
	}
	if !act.CanPass {
		t.Error("起始节点 CanPass 应为 true")
	}
}

func TestWhatCanIDo_FinishedTask(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetTaskInfoFunc = func(ctx context.Context, taskID int) (model.TaskView, error) {
		return model.TaskView{TaskID: 1, IsFinished: 1, ProcInstID: 1}, nil
	}

	act, err := eng.WhatCanIDo(ctx, 1)
	if err != nil {
		t.Fatalf("WhatCanIDo 返回错误: %v", err)
	}
	if act.CanPass || act.CanReject || act.CanFreeRejectToUpstreamNode || act.CanDirectlyToWhoRejectedMe || act.CanRevoke {
		t.Error("已完成任务的所有限制标志应为 false")
	}
}

func TestWhatCanIDo_CosignedNode(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetTaskInfoFunc = func(ctx context.Context, taskID int) (model.TaskView, error) {
		return model.TaskView{
			TaskID:     1,
			IsFinished: 0,
			ProcInstID: 1,
			NodeID:     "Approve",
			IsCosigned: 1,
		}, nil
	}
	repo.GetProcessIDByInstIDFunc = func(ctx context.Context, instID int) (int, error) {
		return 1, nil
	}
	eng.procCache[1] = map[string]model.Node{
		"Approve": {NodeID: "Approve", NodeType: model.TaskNode},
	}
	// isPrevNodeRejected 需要的 mock
	repo.GetPrevNodeBatchCodeFunc = func(ctx context.Context, taskID int) (string, error) {
		return "batch-001", nil
	}
	repo.HasRejectInBatchFunc = func(ctx context.Context, batchCode string) (bool, error) {
		return false, nil
	}

	act, err := eng.WhatCanIDo(ctx, 1)
	if err != nil {
		t.Fatalf("WhatCanIDo 返回错误: %v", err)
	}
	if act.CanDirectlyToWhoRejectedMe {
		t.Error("会签节点 CanDirectlyToWhoRejectedMe 应为 false")
	}
}

func TestWhatCanIDo_NormalNode(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetTaskInfoFunc = func(ctx context.Context, taskID int) (model.TaskView, error) {
		return model.TaskView{
			TaskID:     1,
			IsFinished: 0,
			ProcInstID: 1,
			NodeID:     "Approve",
			IsCosigned: 0,
		}, nil
	}
	repo.GetProcessIDByInstIDFunc = func(ctx context.Context, instID int) (int, error) {
		return 1, nil
	}
	eng.procCache[1] = map[string]model.Node{
		"Approve": {NodeID: "Approve", NodeType: model.TaskNode},
	}
	repo.GetPrevNodeBatchCodeFunc = func(ctx context.Context, taskID int) (string, error) {
		return "batch-001", nil
	}
	repo.HasRejectInBatchFunc = func(ctx context.Context, batchCode string) (bool, error) {
		return false, nil
	}

	act, err := eng.WhatCanIDo(ctx, 1)
	if err != nil {
		t.Fatalf("WhatCanIDo 返回错误: %v", err)
	}
	if !act.CanPass {
		t.Error("普通节点 CanPass 应为 true")
	}
	if !act.CanReject {
		t.Error("普通节点 CanReject 应为 true")
	}
	if !act.CanFreeRejectToUpstreamNode {
		t.Error("普通节点 CanFreeRejectToUpstreamNode 应为 true")
	}
	if act.CanDirectlyToWhoRejectedMe {
		t.Error("无驳回上游时 CanDirectlyToWhoRejectedMe 应为 false")
	}
	if act.CanRevoke {
		t.Error("非起始节点 CanRevoke 应为 false")
	}
}

// ---- TaskTransfer 验证逻辑（事务前部分） ----

func TestTaskTransfer_EmptyUsers(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	err := eng.TaskTransfer(ctx, model.TaskTransferReq{TaskID: 1, Users: nil})
	if err == nil {
		t.Fatal("TaskTransfer 空 users 应返回错误")
	}

	err = eng.TaskTransfer(ctx, model.TaskTransferReq{TaskID: 1, Users: []string{}})
	if err == nil {
		t.Fatal("TaskTransfer 空 users 切片应返回错误")
	}
}

func TestTaskTransfer_AlreadyFinished(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetTaskInfoFunc = func(ctx context.Context, taskID int) (model.TaskView, error) {
		return model.TaskView{TaskID: 1, IsFinished: 1}, nil
	}

	err := eng.TaskTransfer(ctx, model.TaskTransferReq{TaskID: 1, Users: []string{"newuser"}})
	if err == nil {
		t.Fatal("TaskTransfer 已完成任务应返回错误")
	}
}

// ---- isPrevNodeRejected ----

func TestIsPrevNodeRejected_HasReject(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetPrevNodeBatchCodeFunc = func(ctx context.Context, taskID int) (string, error) {
		return "batch-reject", nil
	}
	repo.HasRejectInBatchFunc = func(ctx context.Context, batchCode string) (bool, error) {
		if batchCode != "batch-reject" {
			t.Errorf("batchCode = %q, 期望 %q", batchCode, "batch-reject")
		}
		return true, nil
	}

	taskInfo := model.TaskView{TaskID: 1}
	result, err := eng.isPrevNodeRejected(ctx, taskInfo)
	if err != nil {
		t.Fatalf("isPrevNodeRejected 返回错误: %v", err)
	}
	if !result {
		t.Error("期望返回 true（存在驳回）")
	}
}

func TestIsPrevNodeRejected_NoReject(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetPrevNodeBatchCodeFunc = func(ctx context.Context, taskID int) (string, error) {
		return "batch-pass", nil
	}
	repo.HasRejectInBatchFunc = func(ctx context.Context, batchCode string) (bool, error) {
		return false, nil
	}

	taskInfo := model.TaskView{TaskID: 2}
	result, err := eng.isPrevNodeRejected(ctx, taskInfo)
	if err != nil {
		t.Fatalf("isPrevNodeRejected 返回错误: %v", err)
	}
	if result {
		t.Error("期望返回 false（无驳回）")
	}
}

// ---- getTaskNodeStatus ----

func TestGetTaskNodeStatus(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetTaskNodeStatusFunc = func(ctx context.Context, p repository.TaskNodeStatusParams) (int, int, int, error) {
		return 3, 2, 1, nil
	}

	taskInfo := model.TaskView{ProcInstID: 1, NodeID: "Approve", BatchCode: "batch-001"}
	total, passed, rejected, err := eng.getTaskNodeStatus(ctx, taskInfo)
	if err != nil {
		t.Fatalf("getTaskNodeStatus 返回错误: %v", err)
	}
	if total != 3 || passed != 2 || rejected != 1 {
		t.Errorf("结果 = (%d, %d, %d), 期望 (3, 2, 1)", total, passed, rejected)
	}
}
