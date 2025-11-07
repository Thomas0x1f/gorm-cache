# Quick Start Guide

This guide will help you get started with GORM Cache Plugin quickly.

## 1. Installation

```bash
go get github.com/restayway/gorm-cache
```

## 2. Basic Usage

### In-Memory Cache (5-minute setup)

```go
package main

import (
    "time"
    gormcache "github.com/restayway/gorm-cache"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func main() {
    // Normal GORM setup
    db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    
    // Add cache plugin - THAT'S IT!
    db.Use(gormcache.New(gormcache.Config{
        Adapter: gormcache.NewMemoryAdapter(),
        TTL:     5 * time.Minute,
    }))
    
    // Now all your queries are automatically cached!
    var users []User
    db.Find(&users) // First: from DB
    db.Find(&users) // Second: from cache ⚡
}
```

### Redis Cache (Production Ready)

```go
db.Use(gormcache.New(gormcache.Config{
    Adapter: gormcache.NewRedisAdapter(gormcache.RedisAdapterConfig{
        Addr: "localhost:6379",
    }),
    TTL: 5 * time.Minute,
}))
```

## 3. Common Scenarios

### Scenario 1: Cache Only Specific Models

```go
type User struct { ID uint; Name string }
type Product struct { ID uint; Name string }

db.Use(gormcache.New(gormcache.Config{
    Adapter: gormcache.NewMemoryAdapter(),
    TTL:     5 * time.Minute,
    CacheModels: []interface{}{
        User{},     // ✅ User will be cached
        Product{},  // ✅ Product will be cached
        // Order{} not added, won't be cached
    },
}))
```

### Scenario 2: Skip Cache for Admin

```go
db.Use(gormcache.New(gormcache.Config{
    Adapter: gormcache.NewMemoryAdapter(),
    TTL:     5 * time.Minute,
    SkipCacheCondition: func(db *gorm.DB) bool {
        // Get role from context
        role, _ := db.Statement.Context.Value("user_role").(string)
        return role == "admin" // Skip cache if admin
    },
}))

// Usage
ctx := context.WithValue(ctx, "user_role", "admin")
db.WithContext(ctx).Find(&users) // Cache will be skipped
```

### Scenario 3: Skip Cache for a Single Query

```go
// Method 1: Using scope helper
db.Scopes(gormcache.SkipCache()).Find(&users)

// Method 2: Using context helper (recommended for request-scoped operations)
ctx := gormcache.SkipCacheContext(context.Background())
db.WithContext(ctx).Find(&users)

// Method 3: Lower-level context API
ctx := gormcache.WithSkipCache(context.Background(), true)
db.WithContext(ctx).First(&user, id)
```

### Scenario 4: Auto Cache Clear After Update

```go
db.Use(gormcache.New(gormcache.Config{
    Adapter:            gormcache.NewMemoryAdapter(),
    TTL:                5 * time.Minute,
    InvalidateOnUpdate: true, // ✅ Cache cleared on UPDATE
    InvalidateOnCreate: true, // ✅ Cache cleared on CREATE
    InvalidateOnDelete: true, // ✅ Cache cleared on DELETE
}))

// Example
db.Find(&users)              // Writes to cache
db.Model(&user).Update(...)  // Automatically clears cache
db.Find(&users)              // Reads from DB again, writes to cache
```

## 4. Adapter Selection

| Adapter | When to Use | Pros | Cons |
|---------|-------------|------|------|
| **MemoryAdapter** | Single instance, dev, test | ⚡ Very fast, easy | Lost on restart, uses RAM |
| **RedisAdapter** | Production, multi-instance | 🔄 Persistent, shared | Redis dependency, slightly slower |
| **Custom** | Special needs | 🎯 Full control | You implement it |

## 5. Performance Tips

### ✅ Do's

```go
// 1. Cache only frequently read models
CacheModels: []interface{}{User{}, Product{}}

// 2. Set TTL based on data update frequency
TTL: 10 * time.Minute  // For rarely changed data

// 3. Skip cache for admin/special queries
SkipCacheCondition: func(db *gorm.DB) bool { ... }
```

### ❌ Don'ts

```go
// 1. Don't cache frequently changing data
// ❌ Bad: Dashboard stats changing every second
// ✅ Good: Product list changing once a day

// 2. Don't use very low TTL
TTL: 1 * time.Second  // ❌ Cache becomes pointless

// 3. Don't cache all models
// ❌ Bad: Empty CacheModels (caches everything)
// ✅ Good: Only necessary models
```

## 6. Debugging

```go
// Keep GORM logger enabled
db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})

// You'll see SQL queries:
// - First query: SQL runs
// - Second query: SQL runs but result comes from cache
```

## 7. Example Project Structure

```
myapp/
├── main.go
├── models/
│   ├── user.go
│   └── product.go
├── database/
│   └── cache.go  👈 Cache config here
└── go.mod
```

**database/cache.go:**
```go
package database

import (
    "time"
    gormcache "github.com/restayway/gorm-cache"
)

func SetupCache(db *gorm.DB) error {
    return db.Use(gormcache.New(gormcache.Config{
        Adapter: gormcache.NewRedisAdapter(gormcache.RedisAdapterConfig{
            Addr: os.Getenv("REDIS_ADDR"),
        }),
        TTL: 5 * time.Minute,
        CacheModels: []interface{}{
            models.User{},
            models.Product{},
        },
        InvalidateOnUpdate: true,
        InvalidateOnCreate: true,
        InvalidateOnDelete: true,
    }))
}
```

## 8. FAQ

### How do I know cache is working?

Enable GORM logger. On the second query, SQL query will run but results will be very fast (from cache).

### What happens if Redis connection drops?

Query will work normally from DB, no error.

### Can I manually clear cache?

Yes:
```go
cachePlugin := gormcache.New(...)
// ...
cachePlugin.config.Adapter.Clear(context.Background())
```

### How do I use with multiple DB connections?

Create separate plugin instance for each DB:
```go
plugin1 := gormcache.New(...)
db1.Use(plugin1)

plugin2 := gormcache.New(...)
db2.Use(plugin2)
```

## 9. Next Steps

- [README.md](README.md) - Detailed documentation
- [examples/](examples/) - More examples
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contributing guide

## Need Help?

Ask questions on GitHub Issues: [Issues](https://github.com/restayway/gorm-cache/issues)

---

**Happy Coding! 🚀**
