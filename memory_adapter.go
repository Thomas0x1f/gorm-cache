package gormcache

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

type cacheItem struct {
	value      []byte
	expiration time.Time
}

// MemoryAdapter is an in-memory cache implementation
type MemoryAdapter struct {
	store   map[string]*cacheItem
	mu      sync.RWMutex
	stopCh  chan struct{}
	cleanUp bool
}

// NewMemoryAdapter creates a new in-memory cache adapter
func NewMemoryAdapter() *MemoryAdapter {
	adapter := &MemoryAdapter{
		store:   make(map[string]*cacheItem),
		stopCh:  make(chan struct{}),
		cleanUp: true,
	}

	// Start cleanup goroutine
	go adapter.startCleanup()

	return adapter
}

// Get retrieves a value from memory cache
func (m *MemoryAdapter) Get(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.store[key]
	if !exists {
		return nil, errors.New("key not found")
	}

	// Check if expired
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		return nil, errors.New("key expired")
	}

	return item.value, nil
}

// Set stores a value in memory cache
func (m *MemoryAdapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	item := &cacheItem{
		value: value,
	}

	if ttl > 0 {
		item.expiration = time.Now().Add(ttl)
	}

	m.store[key] = item
	return nil
}

// Delete removes a value from memory cache
func (m *MemoryAdapter) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.store, key)
	return nil
}

// DeletePattern removes all keys matching the pattern
func (m *MemoryAdapter) DeletePattern(ctx context.Context, pattern string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Simple pattern matching: replace * with any characters
	prefix := strings.TrimSuffix(pattern, "*")

	keysToDelete := make([]string, 0)
	for key := range m.store {
		if strings.HasPrefix(key, prefix) || pattern == "*" {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(m.store, key)
	}

	return nil
}

// Clear removes all cached data
func (m *MemoryAdapter) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.store = make(map[string]*cacheItem)
	return nil
}

// Close closes the adapter
func (m *MemoryAdapter) Close() error {
	m.cleanUp = false
	close(m.stopCh)
	return nil
}

// startCleanup periodically removes expired items
func (m *MemoryAdapter) startCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !m.cleanUp {
				return
			}
			m.cleanup()
		case <-m.stopCh:
			return
		}
	}
}

func (m *MemoryAdapter) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, item := range m.store {
		if !item.expiration.IsZero() && now.After(item.expiration) {
			delete(m.store, key)
		}
	}
}
