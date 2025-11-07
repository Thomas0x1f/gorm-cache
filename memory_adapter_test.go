package gormcache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryAdapterSetGet(t *testing.T) {
	adapter := NewMemoryAdapter()
	defer adapter.Close()

	ctx := context.Background()
	key := "test:key"
	value := []byte("test value")

	// Set value
	err := adapter.Set(ctx, key, value, 1*time.Minute)
	if err != nil {
		t.Fatalf("failed to set: %v", err)
	}

	// Get value
	result, err := adapter.Get(ctx, key)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	if string(result) != string(value) {
		t.Errorf("expected '%s', got '%s'", value, result)
	}
}

func TestMemoryAdapterDelete(t *testing.T) {
	adapter := NewMemoryAdapter()
	defer adapter.Close()

	ctx := context.Background()
	key := "test:key"
	value := []byte("test value")

	// Set value
	adapter.Set(ctx, key, value, 1*time.Minute)

	// Delete value
	err := adapter.Delete(ctx, key)
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	// Try to get deleted value
	_, err = adapter.Get(ctx, key)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestMemoryAdapterDeletePattern(t *testing.T) {
	adapter := NewMemoryAdapter()
	defer adapter.Close()

	ctx := context.Background()

	// Set multiple values
	adapter.Set(ctx, "user:1", []byte("value1"), 1*time.Minute)
	adapter.Set(ctx, "user:2", []byte("value2"), 1*time.Minute)
	adapter.Set(ctx, "order:1", []byte("value3"), 1*time.Minute)

	// Delete pattern
	err := adapter.DeletePattern(ctx, "user:*")
	if err != nil {
		t.Fatalf("failed to delete pattern: %v", err)
	}

	// Check user keys are deleted
	_, err = adapter.Get(ctx, "user:1")
	if err == nil {
		t.Error("expected user:1 to be deleted")
	}

	_, err = adapter.Get(ctx, "user:2")
	if err == nil {
		t.Error("expected user:2 to be deleted")
	}

	// Check order key still exists
	_, err = adapter.Get(ctx, "order:1")
	if err != nil {
		t.Error("expected order:1 to exist")
	}
}

func TestMemoryAdapterClear(t *testing.T) {
	adapter := NewMemoryAdapter()
	defer adapter.Close()

	ctx := context.Background()

	// Set multiple values
	adapter.Set(ctx, "key1", []byte("value1"), 1*time.Minute)
	adapter.Set(ctx, "key2", []byte("value2"), 1*time.Minute)

	// Clear all
	err := adapter.Clear(ctx)
	if err != nil {
		t.Fatalf("failed to clear: %v", err)
	}

	// Check all keys are deleted
	_, err = adapter.Get(ctx, "key1")
	if err == nil {
		t.Error("expected key1 to be deleted")
	}

	_, err = adapter.Get(ctx, "key2")
	if err == nil {
		t.Error("expected key2 to be deleted")
	}
}

func TestMemoryAdapterExpiration(t *testing.T) {
	adapter := NewMemoryAdapter()
	defer adapter.Close()

	ctx := context.Background()
	key := "test:key"
	value := []byte("test value")

	// Set value with short TTL
	adapter.Set(ctx, key, value, 100*time.Millisecond)

	// Get value immediately
	_, err := adapter.Get(ctx, key)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Try to get expired value
	_, err = adapter.Get(ctx, key)
	if err == nil {
		t.Error("expected error for expired key, got nil")
	}
}
