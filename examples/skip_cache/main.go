package main

import (
	"context"
	"fmt"
	"time"

	gormcache "github.com/Thomas0x1f/gorm-cache"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Order struct {
	ID     uint
	UserID uint
	Total  float64
}

func main() {
	// Open GORM connection
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the table
	db.AutoMigrate(&Order{})

	// Install cache plugin with custom skip condition
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
	db.Use(cachePlugin)

	// Create test data
	order := Order{UserID: 1, Total: 150.50}
	db.Create(&order)

	fmt.Println("=== Normal User Query (Cached) ===")
	ctx := context.WithValue(context.Background(), "user_role", "user")
	var o1 Order
	db.WithContext(ctx).First(&o1, order.ID)
	fmt.Printf("Order: %+v\n", o1)

	fmt.Println("\n=== Normal User Second Query (From Cache) ===")
	var o2 Order
	db.WithContext(ctx).First(&o2, order.ID)
	fmt.Printf("Order: %+v\n", o2)

	fmt.Println("\n=== Admin User Query (Cache Skipped) ===")
	adminCtx := context.WithValue(context.Background(), "user_role", "admin")
	var o3 Order
	db.WithContext(adminCtx).First(&o3, order.ID)
	fmt.Printf("Order: %+v\n", o3)

	fmt.Println("\n=== Query with SkipCache() Scope Helper ===")
	var o4 Order
	db.Scopes(gormcache.SkipCache()).First(&o4, order.ID)
	fmt.Printf("Order: %+v\n", o4)

	fmt.Println("\n=== Query with SkipCacheContext() Helper ===")
	skipCtx := gormcache.SkipCacheContext(context.Background())
	var o5 Order
	db.WithContext(skipCtx).First(&o5, order.ID)
	fmt.Printf("Order: %+v\n", o5)

	fmt.Println("\n=== Query with WithSkipCache() Helper ===")
	skipCtx2 := gormcache.WithSkipCache(context.Background(), true)
	var o6 Order
	db.WithContext(skipCtx2).First(&o6, order.ID)
	fmt.Printf("Order: %+v\n", o6)

	// Cleanup
	db.Delete(&order)
	cachePlugin.Close()
}
