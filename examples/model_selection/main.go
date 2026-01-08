package main

import (
	"fmt"
	"time"

	gormcache "github.com/Thomas0x1f/gorm-cache"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type CachedModel struct {
	ID   uint
	Name string
}

type NonCachedModel struct {
	ID    uint
	Value string
}

func main() {
	// Open GORM connection
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate tables
	db.AutoMigrate(&CachedModel{}, &NonCachedModel{})

	// Enable cache only for CachedModel
	cachePlugin := gormcache.New(gormcache.Config{
		Adapter: gormcache.NewMemoryAdapter(),
		TTL:     5 * time.Minute,
		CacheModels: []interface{}{
			CachedModel{}, // Only this model will be cached
		},
	})
	db.Use(cachePlugin)

	// Create test data
	cached := CachedModel{Name: "Cached Item"}
	db.Create(&cached)

	nonCached := NonCachedModel{Value: "Non-Cached Item"}
	db.Create(&nonCached)

	fmt.Println("=== CachedModel Query (Will Be Cached) ===")
	var c1 CachedModel
	db.First(&c1, cached.ID)
	fmt.Printf("CachedModel: %+v\n", c1)

	fmt.Println("\n=== CachedModel Second Query (From Cache) ===")
	var c2 CachedModel
	db.First(&c2, cached.ID)
	fmt.Printf("CachedModel: %+v\n", c2)

	fmt.Println("\n=== NonCachedModel Query (Not Cached) ===")
	var n1 NonCachedModel
	db.First(&n1, nonCached.ID)
	fmt.Printf("NonCachedModel: %+v\n", n1)

	fmt.Println("\n=== NonCachedModel Second Query (Still From Database) ===")
	var n2 NonCachedModel
	db.First(&n2, nonCached.ID)
	fmt.Printf("NonCachedModel: %+v\n", n2)

	// Cleanup
	db.Delete(&cached)
	db.Delete(&nonCached)
	cachePlugin.Close()
}
