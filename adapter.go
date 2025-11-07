package gormcache

import (
	"context"
	"time"
)

// Adapter defines the interface for cache storage implementations
type Adapter interface {
	// Get retrieves a value from cache by key
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in cache with the given key and TTL
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a value from cache by key
	Delete(ctx context.Context, key string) error

	// DeletePattern removes all keys matching the pattern
	DeletePattern(ctx context.Context, pattern string) error

	// Clear removes all cached data
	Clear(ctx context.Context) error

	// Close closes the adapter connection
	Close() error
}
