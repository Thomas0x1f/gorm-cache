package gormcache

import (
	"context"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
)

const (
	pluginName = "gorm:cache"
)

// ErrCacheHit 是一个内部使用的 Error，用于在缓存命中时跳过数据库查询
// 这个 Error 会在 afterQueryCallback 中被自动移除，用户不会看到它
type ErrCacheHit struct {
	RowsAffected int64
}

func (e *ErrCacheHit) Error() string {
	return "gorm:cache:hit - this is an internal error and should be cleared by cache plugin"
}

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
	// 如果没有指定序列化器，使用默认的 JSON 序列化器
	if config.Serializer == nil {
		config.Serializer = &JSONSerializer{}
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
	// 如果已经有错误，不处理缓存逻辑
	if db.Error != nil {
		return
	}

	// Skip if cache should be skipped
	if p.config.shouldSkipCache(db) {
		return
	}

	// Skip if model should not be cached
	if !p.config.shouldCacheModel(db) {
		return
	}

	if db.Statement.SQL.Len() == 0 {
		callbacks.BuildQuerySQL(db)
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
		if err := p.config.Serializer.Unmarshal(cachedData, db.Statement.Dest); err == nil {
			// 计算 RowsAffected
			rowsAffected := calculateRowsAffected(db.Statement.Dest)

			// 设置特殊 Error 以跳过数据库查询
			// 注意：此时 db.Error 保证为 nil（函数开头已检查）
			db.Error = &ErrCacheHit{RowsAffected: rowsAffected}
		}
	}
}

// afterQueryCallback is executed after query to store results in cache
func (p *CachePlugin) afterQueryCallback(db *gorm.DB) {
	// 首先检查是否是缓存命中的情况
	if cacheHitErr, ok := db.Error.(*ErrCacheHit); ok {
		// 设置 RowsAffected
		db.RowsAffected = cacheHitErr.RowsAffected

		// 清除 Error，使用户看不到它
		db.Error = nil

		// Already got from cache, no need to store
		return
	}
	// Skip if there was an error
	if db.Error != nil {
		return
	}

	// Skip if cache should be skipped
	if p.config.shouldSkipCache(db) {
		return
	}

	// Skip if model should not be cached
	if !p.config.shouldCacheModel(db) {
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

	// 不缓存空值结果
	if db.RowsAffected == 0 {
		return
	}

	// Serialize result using configured serializer
	cachedData, err := p.config.Serializer.Marshal(db.Statement.Dest)
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

// calculateRowsAffected 计算从缓存恢复的数据的行数
func calculateRowsAffected(dest any) int64 {
	if dest == nil {
		return 0
	}

	reflectValue := reflect.Indirect(reflect.ValueOf(dest))
	switch reflectValue.Kind() {
	case reflect.Slice:
		return int64(reflectValue.Len())
	case reflect.Struct:
		// 单条记录，如果是有效的结构体则返回 1
		return 1
	default:
		return 0
	}
}
