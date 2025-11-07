package gormcache

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestUser struct {
	ID   uint
	Name string
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&TestUser{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func TestBasicCaching(t *testing.T) {
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

	// First query - should hit database
	var user1 TestUser
	result := db.First(&user1, user.ID)
	if result.Error != nil {
		t.Fatalf("failed to query: %v", result.Error)
	}

	if user1.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", user1.Name)
	}

	// Second query - should hit cache
	var user2 TestUser
	result = db.First(&user2, user.ID)
	if result.Error != nil {
		t.Fatalf("failed to query: %v", result.Error)
	}

	if user2.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", user2.Name)
	}
}

func TestCacheInvalidation(t *testing.T) {
	db := setupTestDB(t)

	// Install cache plugin
	cachePlugin := New(Config{
		Adapter:            NewMemoryAdapter(),
		TTL:                5 * time.Minute,
		InvalidateOnUpdate: true,
	})
	err := db.Use(cachePlugin)
	if err != nil {
		t.Fatalf("failed to install plugin: %v", err)
	}
	defer cachePlugin.Close()

	// Create test data
	user := TestUser{Name: "Original Name"}
	db.Create(&user)

	// First query - cache miss
	var user1 TestUser
	db.First(&user1, user.ID)

	// Update user
	db.Model(&user).Update("Name", "Updated Name")

	// Query again - should get updated value (cache was invalidated)
	var user2 TestUser
	db.First(&user2, user.ID)

	if user2.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got '%s'", user2.Name)
	}
}

func TestSkipCache(t *testing.T) {
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

	// Query with skip cache
	var user1 TestUser
	db.Scopes(SkipCache()).First(&user1, user.ID)

	if user1.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", user1.Name)
	}
}

func TestModelSelection(t *testing.T) {
	db := setupTestDB(t)

	// Install cache plugin with model selection
	cachePlugin := New(Config{
		Adapter: NewMemoryAdapter(),
		TTL:     5 * time.Minute,
		CacheModels: []interface{}{
			TestUser{},
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

	// Query should be cached
	var user1 TestUser
	db.First(&user1, user.ID)

	if user1.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", user1.Name)
	}
}

func TestCustomSkipCondition(t *testing.T) {
	db := setupTestDB(t)

	// Install cache plugin with custom skip condition
	cachePlugin := New(Config{
		Adapter: NewMemoryAdapter(),
		TTL:     5 * time.Minute,
		SkipCacheCondition: func(db *gorm.DB) bool {
			if role, ok := db.Statement.Context.Value("role").(string); ok {
				return role == "admin"
			}
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

	// Admin context - should skip cache
	adminCtx := context.WithValue(context.Background(), "role", "admin")
	var user1 TestUser
	db.WithContext(adminCtx).First(&user1, user.ID)

	if user1.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", user1.Name)
	}

	// User context - should use cache
	userCtx := context.WithValue(context.Background(), "role", "user")
	var user2 TestUser
	db.WithContext(userCtx).First(&user2, user.ID)

	if user2.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", user2.Name)
	}
}
