package service

import (
	"context"
	"testing"

	"github.com/Bunny3th/easy-workflow/internal/model"
)

// ---- resolveNodeUsers ----

func TestResolveNodeUsers(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.ResolveVariablesFunc = func(ctx context.Context, instID int, varNames []string) (map[string]string, error) {
		if instID != 1 {
			t.Errorf("instID = %d, 期望 1", instID)
		}
		return map[string]string{
			"$user1": "user1",
			"$user2": "user2",
			"$user3": "user1", // 重复，应被去重
		}, nil
	}

	node := model.Node{
		NodeID:  "Approve",
		UserIDs: []string{"$user1", "$user2", "$user3"},
	}

	users, err := eng.resolveNodeUsers(ctx, 1, node)
	if err != nil {
		t.Fatalf("resolveNodeUsers 返回错误: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("用户数量 = %d, 期望 2（去重后）", len(users))
	}

	userSet := make(map[string]bool)
	for _, u := range users {
		userSet[u] = true
	}
	if !userSet["user1"] || !userSet["user2"] {
		t.Errorf("用户列表 = %v, 应包含 user1 和 user2", users)
	}
}

func TestResolveNodeUsers_SingleUser(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.ResolveVariablesFunc = func(ctx context.Context, instID int, varNames []string) (map[string]string, error) {
		return map[string]string{"$admin": "admin"}, nil
	}

	node := model.Node{
		NodeID:  "Task1",
		UserIDs: []string{"$admin"},
	}

	users, err := eng.resolveNodeUsers(ctx, 1, node)
	if err != nil {
		t.Fatalf("resolveNodeUsers 返回错误: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("用户数量 = %d, 期望 1", len(users))
	}
	if users[0] != "admin" {
		t.Errorf("用户 = %q, 期望 %q", users[0], "admin")
	}
}

func TestResolveNodeUsers_EmptyUserIDs(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.ResolveVariablesFunc = func(ctx context.Context, instID int, varNames []string) (map[string]string, error) {
		return map[string]string{}, nil
	}

	node := model.Node{
		NodeID:  "Task1",
		UserIDs: []string{},
	}

	users, err := eng.resolveNodeUsers(ctx, 1, node)
	if err != nil {
		t.Fatalf("resolveNodeUsers 返回错误: %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("用户数量 = %d, 期望 0", len(users))
	}
}

func TestResolveNodeUsers_ResolveError(t *testing.T) {
	repo := &mockRepo{}
	eng := newTestEngine(repo)
	ctx := context.Background()

	repo.ResolveVariablesFunc = func(ctx context.Context, instID int, varNames []string) (map[string]string, error) {
		return nil, &ErrResolveVariable{}
	}

	node := model.Node{
		NodeID:  "Task1",
		UserIDs: []string{"$undefined_var"},
	}

	_, err := eng.resolveNodeUsers(ctx, 1, node)
	if err == nil {
		t.Fatal("resolveNodeUsers 变量解析失败应返回错误")
	}
}

// ErrResolveVariable 用于测试的变量解析错误
type ErrResolveVariable struct{}

func (e *ErrResolveVariable) Error() string { return "变量解析失败" }
