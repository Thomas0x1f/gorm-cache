package gormcache

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"
)

const (
	pluginName = "gorm:cache"
)

// CachePlugin is a GORM plugin that provides caching functionality
type CachePlugin struct {
	config Config
}

// New creates a new cache plugin with the given configuration
func New(config Config) *CachePlugin {
	if config.Adapter == nil {
		config.Adapter = NewMemoryAdapter()
	}
	if config.TTL == 0 {
		config.TTL = DefaultConfig().TTL
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = DefaultConfig().KeyPrefix
	}

	return &CachePlugin{
		config: config,
	}
}

// Name returns the plugin name
func (p *CachePlugin) Name() string {
	return pluginName
}

// Initialize initializes the plugin with GORM
func (p *CachePlugin) Initialize(db *gorm.DB) error {
	// Register Query callback (for caching SELECT queries)
	err := db.Callback().Query().Before("gorm:query").Register("gorm:cache:query", p.queryCallback)
	if err != nil {
		return err
	}

	// Register After Query callback (for storing results in cache)
	err = db.Callback().Query().After("gorm:query").Register("gorm:cache:after_query", p.afterQueryCallback)
	if err != nil {
		return err
	}

	// Register Create callback (for invalidating cache)
	if p.config.InvalidateOnCreate {
		err = db.Callback().Create().After("gorm:create").Register("gorm:cache:after_create", p.invalidateCallback)
		if err != nil {
			return err
		}
	}

	// Register Update callback (for invalidating cache)
	if p.config.InvalidateOnUpdate {
		err = db.Callback().Update().After("gorm:update").Register("gorm:cache:after_update", p.invalidateCallback)
		if err != nil {
			return err
		}
	}

	// Register Delete callback (for invalidating cache)
	if p.config.InvalidateOnDelete {
		err = db.Callback().Delete().After("gorm:delete").Register("gorm:cache:after_delete", p.invalidateCallback)
		if err != nil {
			return err
		}
	}

	return nil
}

// queryCallback is executed before query to check cache
func (p *CachePlugin) queryCallback(db *gorm.DB) {
	// Skip if cache should be skipped
	if p.config.shouldSkipCache(db) {
		return
	}

	// Skip if model should not be cached
	if !p.config.shouldCacheModel(db) {
		return
	}

	// Skip if not a SELECT query
	if db.Statement.SQL.String() == "" {
		return
	}

	// Generate cache key
	cacheKey := p.config.generateCacheKey(db)

	// Try to get from cache
	ctx := db.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}

	cachedData, err := p.config.Adapter.Get(ctx, cacheKey)
	if err != nil {
		// Cache miss, continue with normal query
		db.Statement.Settings.Store("gorm:cache:key", cacheKey)
		return
	}

	// Cache hit - deserialize and set the result
	if db.Statement.Dest != nil {
		if err := json.Unmarshal(cachedData, db.Statement.Dest); err == nil {
			// Mark that we used cache
			db.Statement.Settings.Store("gorm:cache:hit", true)
			// Skip the actual query
			db.Statement.SkipHooks = true
		}
	}
}

// afterQueryCallback is executed after query to store results in cache
func (p *CachePlugin) afterQueryCallback(db *gorm.DB) {
	// Check if cache was hit
	if v, ok := db.Statement.Settings.Load("gorm:cache:hit"); ok {
		if hit, ok := v.(bool); ok && hit {
			// Already got from cache, no need to store
			return
		}
	}

	// Skip if cache should be skipped
	if p.config.shouldSkipCache(db) {
		return
	}

	// Skip if model should not be cached
	if !p.config.shouldCacheModel(db) {
		return
	}

	// Skip if there was an error
	if db.Error != nil {
		return
	}

	// Get cache key
	cacheKeyVal, ok := db.Statement.Settings.Load("gorm:cache:key")
	if !ok {
		return
	}
	cacheKey, ok := cacheKeyVal.(string)
	if !ok {
		return
	}

	// Serialize result
	cachedData, err := json.Marshal(db.Statement.Dest)
	if err != nil {
		return
	}

	// Store in cache
	ctx := db.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}

	_ = p.config.Adapter.Set(ctx, cacheKey, cachedData, p.config.TTL)
}

// invalidateCallback is executed after create/update/delete to invalidate cache
func (p *CachePlugin) invalidateCallback(db *gorm.DB) {
	// Skip if there was an error
	if db.Error != nil {
		return
	}

	// Skip if model should not be cached (no need to invalidate)
	if !p.config.shouldCacheModel(db) {
		return
	}

	// Get pattern for this model
	pattern := p.config.getModelPattern(db)

	// Delete all cached queries for this model
	ctx := db.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}

	_ = p.config.Adapter.DeletePattern(ctx, pattern)
}

// Close closes the cache adapter
func (p *CachePlugin) Close() error {
	if p.config.Adapter != nil {
		return p.config.Adapter.Close()
	}
	return nil
}
