# GORM Cache Plugin

A flexible and easy-to-use cache plugin for GORM with Redis and in-memory support.

## Features

- 🚀 **Easy Integration**: Seamless integration with GORM plugin system
- 💾 **Multiple Adapter Support**: Redis, In-Memory, and custom adapter support
- 🎯 **Model-Based Caching**: Choose which models to cache
- 🔄 **Auto Invalidation**: Automatic cache clearing on CREATE, UPDATE, DELETE operations
- ⏭️ **Query-Based Skip**: Skip cache for specific queries
- ⚙️ **Customizable**: Custom cache key generation and skip condition functions
- 🔒 **Thread-Safe**: Safe for concurrent operations

## Installation

```bash
go get github.com/restayway/gorm-cache
```

## Quick Start

### Using In-Memory Cache

```go
package main

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    gormcache "github.com/restayway/gorm-cache"
    "time"
)

type User struct {
    ID   uint
    Name string
}

func main() {
    db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

    // Install cache plugin
    cachePlugin := gormcache.New(gormcache.Config{
        Adapter: gormcache.NewMemoryAdapter(),
        TTL:     5 * time.Minute,
    })
    db.Use(cachePlugin)

    // Normal GORM queries - automatically cached
    var user User
    db.First(&user, 1) // First query from database
    db.First(&user, 1) // Second query from cache
}
```

### Using Redis Cache

```go
package main

import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    gormcache "github.com/restayway/gorm-cache"
    "time"
)

func main() {
    db, _ := gorm.Open(postgres.Open("dsn"), &gorm.Config{})

    // Install cache plugin with Redis adapter
    cachePlugin := gormcache.New(gormcache.Config{
        Adapter: gormcache.NewRedisAdapter(gormcache.RedisAdapterConfig{
            Addr:     "localhost:6379",
            Password: "",
            DB:       0,
        }),
        TTL:                5 * time.Minute,
        InvalidateOnUpdate: true,
        InvalidateOnCreate: true,
        InvalidateOnDelete: true,
    })
    db.Use(cachePlugin)
}
```

## API Reference

### Context-Based API

The plugin provides context-based functions for controlling cache behavior:

```go
// Skip cache using context (recommended)
ctx := gormcache.SkipCacheContext(context.Background())
db.WithContext(ctx).Find(&users)

// Lower-level API - set skip cache value
ctx := gormcache.WithSkipCache(context.Background(), true)
db.WithContext(ctx).Find(&users)

// Enable cache (set skip to false)
ctx := gormcache.WithSkipCache(context.Background(), false)
db.WithContext(ctx).Find(&users)
```

### Scope-Based API

The plugin also provides GORM scope helpers:

```go
// Skip cache using scope
db.Scopes(gormcache.SkipCache()).Find(&users)

// Enable cache using scope
db.Scopes(gormcache.EnableCache()).Find(&users)
```

## Advanced Usage

### Cache Specific Models

```go
type User struct {
    ID   uint
    Name string
}

type Order struct {
    ID     uint
    UserID uint
}

cachePlugin := gormcache.New(gormcache.Config{
    Adapter: gormcache.NewMemoryAdapter(),
    TTL:     5 * time.Minute,
    // Only cache User and Order models
    CacheModels: []interface{}{
        User{},
        Order{},
    },
})
db.Use(cachePlugin)
```

### Query-Based Cache Skip

```go
// Method 1: Using SkipCache scope helper
var users []User
db.Scopes(gormcache.SkipCache()).Find(&users)

// Method 2: Using context helper (recommended)
ctx := gormcache.SkipCacheContext(context.Background())
db.WithContext(ctx).Find(&users)

// Method 3: Lower-level context API
ctx := gormcache.WithSkipCache(context.Background(), true)
db.WithContext(ctx).Find(&users)

// Method 4: Custom skip condition
cachePlugin := gormcache.New(gormcache.Config{
    Adapter: gormcache.NewMemoryAdapter(),
    TTL:     5 * time.Minute,
    SkipCacheCondition: func(db *gorm.DB) bool {
        // Skip cache for admin users
        if userRole, ok := db.Statement.Context.Value("user_role").(string); ok {
            return userRole == "admin"
        }
        return false
    },
})
```

### Custom Cache Key Generator

```go
cachePlugin := gormcache.New(gormcache.Config{
    Adapter: gormcache.NewMemoryAdapter(),
    TTL:     5 * time.Minute,
    CacheKeyGenerator: func(db *gorm.DB) string {
        // Custom key generation logic
        return fmt.Sprintf("custom:%s:%v", db.Statement.Table, db.Statement.Vars)
    },
})
```

### Invalidation Settings

```go
cachePlugin := gormcache.New(gormcache.Config{
    Adapter:            gormcache.NewMemoryAdapter(),
    TTL:                5 * time.Minute,
    InvalidateOnUpdate: true,  // Clear cache on UPDATE
    InvalidateOnCreate: false, // Don't clear cache on CREATE
    InvalidateOnDelete: true,  // Clear cache on DELETE
})
```

## Custom Adapter

You can create your own cache adapter:

```go
package main

import (
    "context"
    "time"
    gormcache "github.com/restayway/gorm-cache"
)

type MyCustomAdapter struct {
    // Your custom fields
}

func (a *MyCustomAdapter) Get(ctx context.Context, key string) ([]byte, error) {
    // Get implementation
    return nil, nil
}

func (a *MyCustomAdapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
    // Set implementation
    return nil
}

func (a *MyCustomAdapter) Delete(ctx context.Context, key string) error {
    // Delete implementation
    return nil
}

func (a *MyCustomAdapter) DeletePattern(ctx context.Context, pattern string) error {
    // DeletePattern implementation
    return nil
}

func (a *MyCustomAdapter) Clear(ctx context.Context) error {
    // Clear implementation
    return nil
}

func (a *MyCustomAdapter) Close() error {
    // Close implementation
    return nil
}

// Usage
func main() {
    cachePlugin := gormcache.New(gormcache.Config{
        Adapter: &MyCustomAdapter{},
        TTL:     5 * time.Minute,
    })
}
```

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `Adapter` | `Adapter` | `MemoryAdapter` | Cache storage implementation |
| `TTL` | `time.Duration` | `5 * time.Minute` | Default TTL for cache |
| `CacheModels` | `[]interface{}` | `[]` | Models to cache (empty = all) |
| `InvalidateOnUpdate` | `bool` | `true` | Clear cache on UPDATE |
| `InvalidateOnCreate` | `bool` | `true` | Clear cache on CREATE |
| `InvalidateOnDelete` | `bool` | `true` | Clear cache on DELETE |
| `KeyPrefix` | `string` | `"gorm:cache:"` | Cache key prefix |
| `SkipCacheCondition` | `func(*gorm.DB) bool` | `nil` | Custom condition to skip cache |
| `CacheKeyGenerator` | `func(*gorm.DB) string` | `nil` | Custom cache key generator |

## Performance Tips

1. **TTL Settings**: Set appropriate TTL based on data update frequency
2. **Model Selection**: Only cache frequently read models
3. **Redis vs Memory**: Use Redis for production/multi-instance, memory for single instance/dev/test
4. **Invalidation**: Disable unnecessary invalidation for better performance

## Limitations

- Currently only SELECT queries are cached
- Cache serialization uses JSON
- Memory adapter doesn't persist (data lost on restart)

## Contributing

Contributions are welcome! Please feel free to submit PRs.

## License

MIT License

## Related Links

- [GORM Documentation](https://gorm.io/)
- [GORM Plugin Writing Guide](https://gorm.io/docs/write_plugins.html)
