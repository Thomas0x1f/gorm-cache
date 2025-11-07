package gormcache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Config holds the configuration for the cache plugin
type Config struct {
	// Adapter is the cache storage implementation
	Adapter Adapter

	// TTL is the default time-to-live for cached data
	TTL time.Duration

	// CacheModels defines which models should be cached
	// If empty, all models will be cached
	CacheModels []interface{}

	// InvalidateOnUpdate determines if cache should be cleared on UPDATE operations
	InvalidateOnUpdate bool

	// InvalidateOnCreate determines if cache should be cleared on CREATE operations
	InvalidateOnCreate bool

	// InvalidateOnDelete determines if cache should be cleared on DELETE operations
	InvalidateOnDelete bool

	// KeyPrefix is the prefix for all cache keys
	KeyPrefix string

	// SkipCacheCondition is a function to determine if cache should be skipped for a query
	// Example: func(db *gorm.DB) bool { return db.Statement.Context.Value("skip_cache") == true }
	SkipCacheCondition func(*gorm.DB) bool

	// CacheKeyGenerator allows custom cache key generation
	// If nil, default key generator will be used
	CacheKeyGenerator func(*gorm.DB) string
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		Adapter:            NewMemoryAdapter(),
		TTL:                5 * time.Minute,
		CacheModels:        []interface{}{},
		InvalidateOnUpdate: true,
		InvalidateOnCreate: true,
		InvalidateOnDelete: true,
		KeyPrefix:          "gorm:cache:",
		SkipCacheCondition: nil,
		CacheKeyGenerator:  nil,
	}
}

// shouldCacheModel checks if the model should be cached
func (c *Config) shouldCacheModel(db *gorm.DB) bool {
	// If no models specified, cache all
	if len(c.CacheModels) == 0 {
		return true
	}

	// Check if model is in the cache list
	if db.Statement.Schema == nil {
		return false
	}

	modelType := db.Statement.Schema.ModelType
	for _, cacheModel := range c.CacheModels {
		if fmt.Sprintf("%T", cacheModel) == modelType.String() {
			return true
		}
	}

	return false
}

// shouldSkipCache checks if cache should be skipped
func (c *Config) shouldSkipCache(db *gorm.DB) bool {
	// Check context first
	if skip, ok := getSkipCacheFromContext(db.Statement.Context); ok && skip {
		return true
	}

	// Check custom skip condition
	if c.SkipCacheCondition != nil && c.SkipCacheCondition(db) {
		return true
	}

	// Check if explicitly disabled in the statement
	if v, ok := db.Statement.Settings.Load("gorm:cache:skip"); ok {
		if skip, ok := v.(bool); ok && skip {
			return true
		}
	}

	return false
}

// generateCacheKey generates a cache key for the query
func (c *Config) generateCacheKey(db *gorm.DB) string {
	// Use custom generator if provided
	if c.CacheKeyGenerator != nil {
		return c.KeyPrefix + c.CacheKeyGenerator(db)
	}

	// Default key generation
	key := struct {
		SQL  string
		Vars []interface{}
	}{
		SQL:  db.Statement.SQL.String(),
		Vars: db.Statement.Vars,
	}

	jsonBytes, _ := json.Marshal(key)
	hash := md5.Sum(jsonBytes)

	tableName := "unknown"
	if db.Statement.Schema != nil {
		tableName = db.Statement.Schema.Table
	}

	return c.KeyPrefix + tableName + ":" + hex.EncodeToString(hash[:])
}

// getModelPattern returns the cache key pattern for a model
func (c *Config) getModelPattern(db *gorm.DB) string {
	if db.Statement.Schema == nil {
		return c.KeyPrefix + "*"
	}
	return c.KeyPrefix + db.Statement.Schema.Table + ":*"
}
