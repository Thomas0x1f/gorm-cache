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

type User struct {
	ID    uint
	Name  string
	Email string
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
	db.AutoMigrate(&User{})

	// Install cache plugin
	cachePlugin := gormcache.New(gormcache.Config{
		Adapter: gormcache.NewMemoryAdapter(),
		TTL:     5 * time.Minute,
	})
	db.Use(cachePlugin)

	// Create test data
	user := User{Name: "John Doe", Email: "john@example.com"}
	db.Create(&user)

	fmt.Println("=== First Query (From Database) ===")
	var user1 User
	db.First(&user1, user.ID)
	fmt.Printf("User: %+v\n", user1)

	fmt.Println("\n=== Second Query (From Cache) ===")
	var user2 User
	db.First(&user2, user.ID)
	fmt.Printf("User: %+v\n", user2)

	fmt.Println("\n=== Third Query with SkipCacheContext (From Database) ===")
	ctx := gormcache.SkipCacheContext(context.Background())
	var user3 User
	db.WithContext(ctx).First(&user3, user.ID)
	fmt.Printf("User: %+v\n", user3)

	fmt.Println("\n=== Fourth Query with WithSkipCache(true) (From Database) ===")
	ctx2 := gormcache.WithSkipCache(context.Background(), true)
	var user4 User
	db.WithContext(ctx2).First(&user4, user.ID)
	fmt.Printf("User: %+v\n", user4)

	fmt.Println("\n=== Fifth Query with WithSkipCache(false) (From Cache) ===")
	ctx3 := gormcache.WithSkipCache(context.Background(), false)
	var user5 User
	db.WithContext(ctx3).First(&user5, user.ID)
	fmt.Printf("User: %+v\n", user5)

	fmt.Println("\n=== Sixth Query Normal (From Cache) ===")
	var user6 User
	db.First(&user6, user.ID)
	fmt.Printf("User: %+v\n", user6)

	// Cleanup
	db.Delete(&user)
	cachePlugin.Close()
}
