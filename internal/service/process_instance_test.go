package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/repository"
)

// ---- GetInstanceInfo ----

func TestGetInstanceInfo(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	expected := model.InstanceView{
		ProcInstID:    1,
		ProcID:        10,
		ProcName:      "请假流程",
		BusinessID:    "BIZ-001",
		Starter:       "user1",
		CurrentNodeID: "Approve",
		Status:        0,
	}
	repo.GetInstanceInfoFunc = func(ctx context.Context, instID int) (model.InstanceView, error) {
		if instID != 1 {
			t.Errorf("instID = %d, 期望 1", instID)
		}
		return expected, nil
	}

	result, err := eng.GetInstanceInfo(ctx, 1)
	if err != nil {
		t.Fatalf("GetInstanceInfo 返回错误: %v", err)
	}
	if result.ProcName != "请假流程" {
		t.Errorf("ProcName = %q, 期望 %q", result.ProcName, "请假流程")
	}
	if result.Starter != "user1" {
		t.Errorf("Starter = %q, 期望 %q", result.Starter, "user1")
	}
	if result.CurrentNodeID != "Approve" {
		t.Errorf("CurrentNodeID = %q, 期望 %q", result.CurrentNodeID, "Approve")
	}
}

func TestGetInstanceInfo_NotFound(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetInstanceInfoFunc = func(ctx context.Context, instID int) (model.InstanceView, error) {
		return model.InstanceView{}, nil
	}

	result, err := eng.GetInstanceInfo(ctx, 999)
	if err != nil {
		t.Fatalf("GetInstanceInfo 返回错误: %v", err)
	}
	if result.ProcInstID != 0 {
		t.Error("不存在的实例应返回零值 InstanceView")
	}
}

// ---- GetInstanceStartByUser ----

func TestGetInstanceStartByUser(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	expected := []model.InstanceView{
		{ProcInstID: 1, ProcName: "请假流程", Starter: "user1"},
		{ProcInstID: 2, ProcName: "报销流程", Starter: "user1"},
	}
	repo.ListInstanceStartByUserFunc = func(ctx context.Context, p repository.ListInstByUserParams) ([]model.InstanceView, error) {
		if p.UserID != "user1" {
			t.Errorf("userID = %q, 期望 %q", p.UserID, "user1")
		}
		if p.Offset != 0 || p.Limit != 10 {
			t.Errorf("offset=%d limit=%d, 期望 offset=0 limit=10", p.Offset, p.Limit)
		}
		return expected, nil
	}
	repo.CountInstanceStartByUserFunc = func(ctx context.Context, p repository.CountByUserParams) (int64, error) {
		return 2, nil
	}

	result, err := eng.GetInstanceStartByUser(ctx, model.InstanceListByUserParams{UserID: "user1", ProcessName: "", PageNo: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("GetInstanceStartByUser 返回错误: %v", err)
	}
	if len(result.Data) != 2 {
		t.Fatalf("返回数量 = %d, 期望 2", len(result.Data))
	}
	if result.Count != 2 {
		t.Errorf("Count = %d, 期望 2", result.Count)
	}
}

func TestGetInstanceStartByUser_WithProcessName(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.ListInstanceStartByUserFunc = func(ctx context.Context, p repository.ListInstByUserParams) ([]model.InstanceView, error) {
		if p.ProcessName != "请假流程" {
			t.Errorf("processName = %q, 期望 %q", p.ProcessName, "请假流程")
		}
		return []model.InstanceView{{ProcInstID: 1}}, nil
	}
	repo.CountInstanceStartByUserFunc = func(ctx context.Context, p repository.CountByUserParams) (int64, error) {
		return 1, nil
	}

	result, err := eng.GetInstanceStartByUser(ctx, model.InstanceListByUserParams{UserID: "user1", ProcessName: "请假流程", PageNo: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("GetInstanceStartByUser 返回错误: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("返回数量 = %d, 期望 1", len(result.Data))
	}
}

// ---- InstanceRevoke ----

// 注意：InstanceRevoke 内部使用了 db.Transaction，在无数据库连接的测试中无法完整执行。
// 以下仅测试 force=false 时在 db 调用前的逻辑。

func TestInstanceRevoke_Force_RepoError(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	// force=true 时跳过节点位置检查，但 GetProcessIDByInstID 出错时应返回错误
	repo.GetProcessIDByInstIDFunc = func(ctx context.Context, instID int) (int, error) {
		return 0, errors.New("流程实例不存在")
	}

	err := eng.InstanceRevoke(ctx, model.InstanceRevokeParams{InstanceID: 1, Force: true, RevokeUserID: "user1"})
	if err == nil {
		t.Fatal("InstanceRevoke GetProcessIDByInstID 返回错误时应传播")
	}
}

// ---- getProcCache ----

func TestGetProcCache_CacheHit(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	nodes := map[string]model.Node{
		"Start":   {NodeID: "Start", NodeType: model.RootNode},
		"Approve": {NodeID: "Approve", NodeType: model.TaskNode},
	}
	eng.procCache[1] = nodes

	result, err := eng.getProcCache(ctx, 1)
	if err != nil {
		t.Fatalf("getProcCache 返回错误: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("缓存节点数量 = %d, 期望 2", len(result))
	}
	if result["Start"].NodeType != model.RootNode {
		t.Error("Start 节点类型不正确")
	}
}

func TestGetProcCache_CacheMiss(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetProcessResourceFunc = func(ctx context.Context, procID int) (string, error) {
		return `{"ProcessName":"请假","Source":"OA","Nodes":[{"NodeID":"Start","NodeType":0}]}`, nil
	}

	result, err := eng.getProcCache(ctx, 2)
	if err != nil {
		t.Fatalf("getProcCache 返回错误: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("缓存节点数量 = %d, 期望 1", len(result))
	}
	// 验证缓存已填充
	if _, ok := eng.procCache[2]; !ok {
		t.Error("缓存未命中时应填充缓存")
	}
}

// ---- getInstanceNode ----

func TestGetInstanceNode(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetProcessIDByInstIDFunc = func(ctx context.Context, instID int) (int, error) {
		return 1, nil
	}
	repo.GetProcessResourceFunc = func(ctx context.Context, procID int) (string, error) {
		return `{"ProcessName":"请假","Source":"OA","Nodes":[{"NodeID":"Approve","NodeName":"审批","NodeType":1}]}`, nil
	}

	node, err := eng.getInstanceNode(ctx, 1, "Approve")
	if err != nil {
		t.Fatalf("getInstanceNode 返回错误: %v", err)
	}
	if node.NodeID != "Approve" {
		t.Errorf("NodeID = %q, 期望 %q", node.NodeID, "Approve")
	}
	if node.NodeName != "审批" {
		t.Errorf("NodeName = %q, 期望 %q", node.NodeName, "审批")
	}
}

func TestGetInstanceNode_NotFound(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetProcessIDByInstIDFunc = func(ctx context.Context, instID int) (int, error) {
		return 1, nil
	}
	repo.GetProcessResourceFunc = func(ctx context.Context, procID int) (string, error) {
		return `{"ProcessName":"请假","Source":"OA","Nodes":[{"NodeID":"Start","NodeType":0}]}`, nil
	}

	_, err := eng.getInstanceNode(ctx, 1, "NotExist")
	if err == nil {
		t.Fatal("getInstanceNode 不存在的节点应返回错误")
	}
}
