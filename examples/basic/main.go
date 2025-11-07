package main

import (
	"fmt"
	"time"

	gormcache "github.com/restayway/gorm-cache"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	ID        uint
	Name      string
	Email     string
	CreatedAt time.Time
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
		Adapter:            gormcache.NewMemoryAdapter(),
		TTL:                5 * time.Minute,
		InvalidateOnUpdate: true,
		InvalidateOnCreate: true,
		InvalidateOnDelete: true,
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

	fmt.Println("\n=== Update User ===")
	db.Model(&user).Update("Name", "Jane Doe")

	fmt.Println("\n=== Third Query (From Database - Cache Cleared) ===")
	var user3 User
	db.First(&user3, user.ID)
	fmt.Printf("User: %+v\n", user3)

	fmt.Println("\n=== Fourth Query (From Cache) ===")
	var user4 User
	db.First(&user4, user.ID)
	fmt.Printf("User: %+v\n", user4)

	// Cleanup
	db.Delete(&user)
	cachePlugin.Close()
}
