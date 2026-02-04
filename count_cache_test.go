package gormcache

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestCountQueryCache 测试 COUNT 查询的缓存
func TestCountQueryCache(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// 迁移
	db.AutoMigrate(&TestUser{})

	// 安装缓存插件（使用 MsgPack 序列化器）
	cachePlugin := New(Config{
		Adapter:    NewMemoryAdapter(),
		TTL:        5 * time.Minute,
		Serializer: &MsgPackSerializer{},
	})
	err = db.Use(cachePlugin)
	if err != nil {
		t.Fatalf("failed to install plugin: %v", err)
	}
	defer cachePlugin.Close()

	// 创建测试数据
	user := TestUser{Name: "Test User"}
	db.Create(&user)

	// 第一次 COUNT 查询 - 从数据库获取
	var count1 int64
	result1 := db.Model(&TestUser{}).Count(&count1)
	if result1.Error != nil {
		t.Fatalf("first count query failed: %v", result1.Error)
	}
	t.Logf("First count query - Count: %d, RowsAffected: %d", count1, result1.RowsAffected)

	if count1 != 1 {
		t.Errorf("Expected count 1, got %d", count1)
	}

	// 第二次 COUNT 查询 - 应该从缓存获取
	var count2 int64
	result2 := db.Model(&TestUser{}).Count(&count2)
	if result2.Error != nil {
		t.Fatalf("second count query failed: %v", result2.Error)
	}
	t.Logf("Second count query - Count: %d, RowsAffected: %d", count2, result2.RowsAffected)

	// 验证缓存的结果
	if count2 != 1 {
		t.Errorf("Expected cached count 1, got %d", count2)
	}
}

// TestCountQueryCacheWithCondition 测试带条件的 COUNT 查询
func TestCountQueryCacheWithCondition(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// 迁移
	db.AutoMigrate(&TestUser{})

	// 安装缓存插件（使用 MsgPack 序列化器）
	cachePlugin := New(Config{
		Adapter:    NewMemoryAdapter(),
		TTL:        5 * time.Minute,
		Serializer: &MsgPackSerializer{},
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

	// 第一次查询 - 从数据库获取
	var count1 int64
	result1 := db.Model(&TestUser{}).Where("name LIKE ?", "User%").Count(&count1)
	if result1.Error != nil {
		t.Fatalf("first query failed: %v", result1.Error)
	}
	t.Logf("First query - Count: %d, RowsAffected: %d", count1, result1.RowsAffected)

	// 第二次查询 - 从缓存获取
	var count2 int64
	result2 := db.Model(&TestUser{}).Where("name LIKE ?", "User%").Count(&count2)
	if result2.Error != nil {
		t.Fatalf("second query failed: %v", result2.Error)
	}
	t.Logf("Second query - Count: %d, RowsAffected: %d", count2, result2.RowsAffected)

	// 验证结果
	if count2 != 2 {
		t.Errorf("Expected count 2, got %d", count2)
	}
}
