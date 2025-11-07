package gormcache

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestSkipCacheContext(t *testing.T) {
	db := setupTestDB(t)

	// Install cache plugin
	cachePlugin := New(Config{
		Adapter: NewMemoryAdapter(),
		TTL:     5 * time.Minute,
	})
	err := db.Use(cachePlugin)
	if err != nil {
		t.Fatalf("failed to install plugin: %v", err)
	}
	defer cachePlugin.Close()

	// Create test data
	user := TestUser{Name: "Test User"}
	db.Create(&user)

	// First query - cache miss
	var user1 TestUser
	db.First(&user1, user.ID)

	// Second query with skip cache context
	ctx := SkipCacheContext(context.Background())
	var user2 TestUser
	result := db.WithContext(ctx).First(&user2, user.ID)
	if result.Error != nil {
		t.Fatalf("failed to query: %v", result.Error)
	}

	if user2.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", user2.Name)
	}
}

func TestContextSkipCachePriority(t *testing.T) {
	db := setupTestDB(t)

	// Install cache plugin with custom skip condition
	cachePlugin := New(Config{
		Adapter: NewMemoryAdapter(),
		TTL:     5 * time.Minute,
		SkipCacheCondition: func(db *gorm.DB) bool {
			// This should not be reached if context skip is set
			return false
		},
	})
	err := db.Use(cachePlugin)
	if err != nil {
		t.Fatalf("failed to install plugin: %v", err)
	}
	defer cachePlugin.Close()

	// Create test data
	user := TestUser{Name: "Test User"}
	db.Create(&user)

	// Context skip should take priority
	ctx := SkipCacheContext(context.Background())
	var user1 TestUser
	result := db.WithContext(ctx).First(&user1, user.ID)
	if result.Error != nil {
		t.Fatalf("failed to query: %v", result.Error)
	}

	if user1.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", user1.Name)
	}
}

func TestWithSkipCacheFunction(t *testing.T) {
	db := setupTestDB(t)

	// Install cache plugin
	cachePlugin := New(Config{
		Adapter: NewMemoryAdapter(),
		TTL:     5 * time.Minute,
	})
	err := db.Use(cachePlugin)
	if err != nil {
		t.Fatalf("failed to install plugin: %v", err)
	}
	defer cachePlugin.Close()

	// Create test data
	user := TestUser{Name: "Test User"}
	db.Create(&user)

	// Use WithSkipCache function
	ctx := WithSkipCache(context.Background(), true)
	var user1 TestUser
	result := db.WithContext(ctx).First(&user1, user.ID)
	if result.Error != nil {
		t.Fatalf("failed to query: %v", result.Error)
	}

	if user1.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", user1.Name)
	}
}
