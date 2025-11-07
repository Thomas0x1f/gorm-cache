package gormcache

import "gorm.io/gorm"

// SkipCache is a scope helper function to skip cache for a specific query
// Usage: db.Scopes(gormcache.SkipCache()).Find(&users)
func SkipCache() func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db.Statement.Settings.Store("gorm:cache:skip", true)
		return db
	}
}

// EnableCache is a scope helper function to explicitly enable cache for a query
// Usage: db.Scopes(gormcache.EnableCache()).Find(&users)
func EnableCache() func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db.Statement.Settings.Store("gorm:cache:skip", false)
		return db
	}
}
