package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/Bunny3th/easy-workflow/internal/entity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const testDSN = "root:root@tcp(127.0.0.1:3306)/easy_workflow_v2?charset=utf8mb4&parseTime=True&loc=Local"

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(mysql.Open(testDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("连接数据库失败: %v", err)
	}
	return db
}

// insertTestData 插入测试数据（proc_def + proc_inst），返回 instID 和 procID。
// 如果 proc_inst 表已有数据则直接返回第一条记录。
func insertTestData(t *testing.T, db *gorm.DB) (instID int, procID int) {
	t.Helper()

	// 检查是否已有数据
	var count int64
	db.Table("proc_inst").Count(&count)
	if count > 0 {
		if err := db.Table("proc_inst").Select("id, proc_id").Row().Scan(&instID, &procID); err != nil {
			t.Fatalf("查询 proc_inst 失败: %v", err)
		}
		t.Logf("proc_inst 已有数据，使用 instID=%d, proc_id=%d", instID, procID)
		return
	}

	// 插入 proc_def
	if err := db.Table("proc_def").Create(map[string]any{
		"name": "__test_flow", "version": 1, "resource": `{"nodes":[]}`,
		"user_id": "test_user", "source": "unit_test", "create_time": entity.Now(),
	}).Error; err != nil {
		t.Fatalf("插入 proc_def 失败: %v", err)
	}
	db.Table("proc_def").Select("id").Where("name = ?", "__test_flow").Scan(&procID)
	t.Logf("插入 proc_def: id=%d", procID)

	// 插入 proc_inst
	if err := db.Table("proc_inst").Create(map[string]any{
		"proc_id": procID, "proc_version": 1, "business_id": "biz_001",
		"starter": "test_user", "current_node_id": "start_node",
		"create_time": entity.Now(), "status": 0,
	}).Error; err != nil {
		t.Fatalf("插入 proc_inst 失败: %v", err)
	}
	db.Table("proc_inst").Select("id").Where("business_id = ?", "biz_001").Scan(&instID)
	t.Logf("插入 proc_inst: id=%d, proc_id=%d", instID, procID)

	return
}

// cleanTestData 清理测试数据。
func cleanTestData(t *testing.T, db *gorm.DB) {
	t.Helper()
	db.Exec("DELETE FROM proc_inst WHERE starter = 'test_user' AND business_id = 'biz_001'")
	db.Exec("DELETE FROM proc_def WHERE name = '__test_flow' AND source = 'unit_test'")
}

func TestFlowRepo_GetProcessIDByInstID(t *testing.T) {
	db := setupTestDB(t)
	instID, procID := insertTestData(t, db)
	defer cleanTestData(t, db)

	repo := NewFlowRepo(db)
	ctx := context.Background()

	t.Logf("测试: instID=%d, 期望 proc_id=%d", instID, procID)

	got, err := repo.GetProcessIDByInstID(ctx, instID)
	if err != nil {
		t.Fatalf("GetProcessIDByInstID(%d) 返回错误: %v", instID, err)
	}
	if got != procID {
		t.Errorf("GetProcessIDByInstID(%d) = %d, 期望 %d", instID, got, procID)
	} else {
		fmt.Printf("GetProcessIDByInstID(%d) = %d ✓\n", instID, got)
	}
}

func TestFlowRepo_GetProcessIDByInstID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFlowRepo(db)
	ctx := context.Background()

	got, err := repo.GetProcessIDByInstID(ctx, -1)
	if err == nil {
		if got != 0 {
			t.Errorf("不存在的 instID 应返回 0, 实际返回 %d", got)
		} else {
			fmt.Println("不存在的 instID 返回 0 ✓（Scan 零值语义）")
		}
	} else {
		t.Logf("不存在的 instID 返回错误: %v", err)
	}
}
