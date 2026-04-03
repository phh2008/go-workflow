package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"
	"testing"

	"github.com/Bunny3th/easy-workflow/internal/entity"
	"github.com/Bunny3th/easy-workflow/internal/model"
	"github.com/Bunny3th/easy-workflow/internal/repository"
	"gorm.io/gorm"
)

// newTestEngine 创建用于测试的 Engine 实例（无数据库连接）。
func newTestEngine(repo repository.Repository) *Engine {
	return &Engine{
		logger:         slog.Default(),
		repo:           repo,
		eventPool:      make(map[string]*eventMethod),
		ignoreEventErr: false,
		procCache:      make(map[int]map[string]model.Node),
		scheduledTasks: make(map[string]*SchedulerTask),
		expressionEval: NewExpressionEvaluator(),
	}
}

// ---- ProcessParse ----

func TestProcessParse_ValidJSON(t *testing.T) {
	eng := newTestEngine(&mockRepo{})
	ctx := context.Background()

	resource := `{
		"ProcessName": "请假流程",
		"Source": "OA系统",
		"RevokeEvents": ["OnRevoke"],
		"Nodes": [
			{
				"NodeID": "Start",
				"NodeName": "开始",
				"NodeType": 0,
				"PrevNodeIDs": [],
				"UserIDs": ["$starter"]
			},
			{
				"NodeID": "End",
				"NodeName": "结束",
				"NodeType": 3,
				"PrevNodeIDs": ["Approve"]
			}
		]
	}`

	proc, err := eng.ProcessParse(ctx, resource)
	if err != nil {
		t.Fatalf("ProcessParse 返回错误: %v", err)
	}
	if proc.ProcessName != "请假流程" {
		t.Errorf("ProcessName = %q, 期望 %q", proc.ProcessName, "请假流程")
	}
	if proc.Source != "OA系统" {
		t.Errorf("Source = %q, 期望 %q", proc.Source, "OA系统")
	}
	if len(proc.Nodes) != 2 {
		t.Fatalf("Nodes 数量 = %d, 期望 2", len(proc.Nodes))
	}
	if proc.Nodes[0].NodeID != "Start" {
		t.Errorf("第一个节点 NodeID = %q, 期望 %q", proc.Nodes[0].NodeID, "Start")
	}
	if proc.Nodes[0].NodeType != model.RootNode {
		t.Errorf("第一个节点 NodeType = %d, 期望 %d", proc.Nodes[0].NodeType, model.RootNode)
	}
	if len(proc.RevokeEvents) != 1 || proc.RevokeEvents[0] != "OnRevoke" {
		t.Errorf("RevokeEvents = %v, 期望 [OnRevoke]", proc.RevokeEvents)
	}
}

func TestProcessParse_InvalidJSON(t *testing.T) {
	eng := newTestEngine(&mockRepo{})
	ctx := context.Background()

	_, err := eng.ProcessParse(ctx, "not json at all")
	if err == nil {
		t.Fatal("ProcessParse 应返回错误，但返回 nil")
	}
}

// ---- ProcessSave 验证逻辑（事务前部分） ----

func TestProcessSave_EmptyName(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	resource := `{"ProcessName":"","Source":"OA","Nodes":[]}`
	_, err := eng.ProcessSave(ctx, resource, "user1")
	if err == nil {
		t.Fatal("ProcessSave 空名称应返回错误")
	}
}

func TestProcessSave_EmptySource(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	resource := `{"ProcessName":"请假","Source":"","Nodes":[]}`
	_, err := eng.ProcessSave(ctx, resource, "user1")
	if err == nil {
		t.Fatal("ProcessSave 空来源应返回错误")
	}
}

func TestProcessSave_EmptyCreateUserID(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	resource := `{"ProcessName":"请假","Source":"OA","Nodes":[]}`
	_, err := eng.ProcessSave(ctx, resource, "")
	if err == nil {
		t.Fatal("ProcessSave 空创建人ID应返回错误")
	}
}

func TestProcessSave_InvalidJSON(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	_, err := eng.ProcessSave(ctx, "invalid", "user1")
	if err == nil {
		t.Fatal("ProcessSave 无效 JSON 应返回错误")
	}
}

// ---- GetProcessDefine ----

func TestGetProcessDefine(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	expectedProc := model.Process{
		ProcessName: "请假流程",
		Source:      "OA系统",
		Nodes: []model.Node{
			{NodeID: "Start", NodeName: "开始", NodeType: model.RootNode},
		},
	}
	resourceBytes, _ := json.Marshal(expectedProc)

	repo.GetProcessResourceFunc = func(ctx context.Context, procID int) (string, error) {
		if procID != 1 {
			t.Errorf("procID = %d, 期望 1", procID)
		}
		return string(resourceBytes), nil
	}

	proc, err := eng.GetProcessDefine(ctx, 1)
	if err != nil {
		t.Fatalf("GetProcessDefine 返回错误: %v", err)
	}
	if proc.ProcessName != "请假流程" {
		t.Errorf("ProcessName = %q, 期望 %q", proc.ProcessName, "请假流程")
	}
	if len(proc.Nodes) != 1 || proc.Nodes[0].NodeID != "Start" {
		t.Errorf("Nodes 不符合预期: %v", proc.Nodes)
	}
}

func TestGetProcessDefine_EmptyResource(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetProcessResourceFunc = func(ctx context.Context, procID int) (string, error) {
		return "", nil
	}

	_, err := eng.GetProcessDefine(ctx, 999)
	if err == nil {
		t.Fatal("GetProcessDefine 空资源应返回错误")
	}
}

func TestGetProcessDefine_RepoError(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.GetProcessResourceFunc = func(ctx context.Context, procID int) (string, error) {
		return "", gorm.ErrRecordNotFound
	}

	_, err := eng.GetProcessDefine(ctx, 1)
	if err == nil {
		t.Fatal("GetProcessDefine 应传播 repo 错误")
	}
}

// ---- GetProcessList ----

func TestGetProcessList(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	expectedDefs := []entity.ProcDef{
		{ID: 1, Name: "请假流程", Source: "OA"},
		{ID: 2, Name: "报销流程", Source: "OA"},
	}
	repo.ListProcessDefFunc = func(ctx context.Context, source string) ([]entity.ProcDef, error) {
		if source != "OA" {
			t.Errorf("source = %q, 期望 %q", source, "OA")
		}
		return expectedDefs, nil
	}

	defs, err := eng.GetProcessList(ctx, "OA")
	if err != nil {
		t.Fatalf("GetProcessList 返回错误: %v", err)
	}
	if len(defs) != 2 {
		t.Fatalf("返回数量 = %d, 期望 2", len(defs))
	}
	if defs[0].Name != "请假流程" {
		t.Errorf("defs[0].Name = %q, 期望 %q", defs[0].Name, "请假流程")
	}
}

func TestGetProcessList_Empty(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.ListProcessDefFunc = func(ctx context.Context, source string) ([]entity.ProcDef, error) {
		return nil, nil
	}

	defs, err := eng.GetProcessList(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetProcessList 返回错误: %v", err)
	}
	if len(defs) != 0 {
		t.Errorf("返回数量 = %d, 期望 0", len(defs))
	}
}

// ---- nodesToExecutions ----

func TestNodesToExecutions_SinglePrev(t *testing.T) {
	nodes := []model.Node{
		{NodeID: "A", NodeName: "节点A", NodeType: model.TaskNode, PrevNodeIDs: []string{"Start"}, IsCosigned: 0},
	}
	executions := nodesToExecutions(1, 1, nodes)
	if len(executions) != 1 {
		t.Fatalf("executions 数量 = %d, 期望 1", len(executions))
	}
	if executions[0].NodeID != "A" {
		t.Errorf("NodeID = %q, 期望 %q", executions[0].NodeID, "A")
	}
	if executions[0].PrevNodeID != "Start" {
		t.Errorf("PrevNodeID = %q, 期望 %q", executions[0].PrevNodeID, "Start")
	}
}

func TestNodesToExecutions_NoPrev(t *testing.T) {
	nodes := []model.Node{
		{NodeID: "Start", NodeName: "开始", NodeType: model.RootNode, PrevNodeIDs: nil, IsCosigned: 0},
	}
	executions := nodesToExecutions(1, 1, nodes)
	if len(executions) != 1 {
		t.Fatalf("executions 数量 = %d, 期望 1", len(executions))
	}
	if executions[0].PrevNodeID != "" {
		t.Errorf("PrevNodeID = %q, 期望空字符串", executions[0].PrevNodeID)
	}
}

func TestNodesToExecutions_MultiplePrev(t *testing.T) {
	nodes := []model.Node{
		{NodeID: "GW", NodeName: "网关", NodeType: model.GateWayNode, PrevNodeIDs: []string{"A", "B", "C"}, IsCosigned: 0},
	}
	executions := nodesToExecutions(1, 1, nodes)
	if len(executions) != 3 {
		t.Fatalf("executions 数量 = %d, 期望 3", len(executions))
	}
	prevNodeIDs := []string{executions[0].PrevNodeID, executions[1].PrevNodeID, executions[2].PrevNodeID}
	for _, expected := range []string{"A", "B", "C"} {
		found := slices.Contains(prevNodeIDs, expected)
		if !found {
			t.Errorf("未找到 PrevNodeID = %q", expected)
		}
	}
}
