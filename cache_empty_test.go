package gormcache

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestEmptyResultsNotCached 测试空结果不被缓存
func TestEmptyResultsNotCached(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// 迁移
	db.AutoMigrate(&TestUser{})

	// 安装缓存插件
	cachePlugin := New(Config{
		Adapter: NewMemoryAdapter(),
		TTL:     5 * time.Minute,
	})
	err = db.Use(cachePlugin)
	if err != nil {
		t.Fatalf("failed to install plugin: %v", err)
	}
	defer cachePlugin.Close()

	// 第一次查询 - 数据库为空，应该不缓存
	var users1 []TestUser
	result1 := db.Find(&users1)
	if result1.Error != nil {
		t.Fatalf("first query failed: %v", result1.Error)
	}
	t.Logf("First query - RowsAffected: %d, Length: %d", result1.RowsAffected, len(users1))

	if len(users1) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(users1))
	}

	// 添加数据到数据库
	user := TestUser{Name: "New User"}
	db.Create(&user)

	// 第二次查询 - 应该从数据库获取，而不是返回缓存的空结果
	var users2 []TestUser
	result2 := db.Find(&users2)
	if result2.Error != nil {
		t.Fatalf("second query failed: %v", result2.Error)
	}
	t.Logf("Second query - RowsAffected: %d, Length: %d", result2.RowsAffected, len(users2))

	// 应该查询到新添加的用户
	if len(users2) != 1 {
		t.Errorf("Expected 1 user after insert, got %d", len(users2))
	}
	if len(users2) > 0 && users2[0].Name != "New User" {
		t.Errorf("Expected user name 'New User', got '%s'", users2[0].Name)
	}
}

// TestNonEmptyResultsCached 测试非空结果仍然被正常缓存
func TestNonEmptyResultsCached(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// 迁移
	db.AutoMigrate(&TestUser{})

	// 安装缓存插件
	cachePlugin := New(Config{
		Adapter: NewMemoryAdapter(),
		TTL:     5 * time.Minute,
	})
	err = db.Use(cachePlugin)
	if err != nil {
		t.Fatalf("failed to install plugin: %v", err)
	}
	defer cachePlugin.Close()

	// 创建测试数据
	users := []TestUser{
		{Name: "User 1"},
		{Name: "User 2"},
	}
	for _, u := range users {
		db.Create(&u)
	}

	// 第一次查询 - 应该缓存
	var users1 []TestUser
	result1 := db.Find(&users1)
	if result1.Error != nil {
		t.Fatalf("first query failed: %v", result1.Error)
	}
	t.Logf("First query - RowsAffected: %d, Length: %d", result1.RowsAffected, len(users1))

	// 清空数据库（但缓存还在）
	db.Unscoped().Where("1 = 1").Delete(&TestUser{})

	// 第二次查询 - 应该从缓存获取（即使数据库已清空）
	var users2 []TestUser
	result2 := db.Find(&users2)
	if result2.Error != nil {
		t.Fatalf("second query failed: %v", result2.Error)
	}
	t.Logf("Second query - RowsAffected: %d, Length: %d", result2.RowsAffected, len(users2))

	// 验证从缓存获取的数据
	if len(users2) != 2 {
		t.Errorf("Expected 2 users from cache, got %d", len(users2))
	}
}
