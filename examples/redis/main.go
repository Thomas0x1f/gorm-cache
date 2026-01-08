package main

import (
	"fmt"
	"time"

	gormcache "github.com/Thomas0x1f/gorm-cache"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Product struct {
	ID    uint
	Name  string
	Price float64
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
	db.AutoMigrate(&Product{})

	// Install Redis cache plugin
	cachePlugin := gormcache.New(gormcache.Config{
		Adapter: gormcache.NewRedisAdapter(gormcache.RedisAdapterConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}),
		TTL:                10 * time.Minute,
		InvalidateOnUpdate: true,
		InvalidateOnCreate: true,
		InvalidateOnDelete: true,
	})
	db.Use(cachePlugin)

	// Create test data
	product := Product{Name: "Laptop", Price: 999.99}
	db.Create(&product)

	fmt.Println("=== First Query (From Database) ===")
	var p1 Product
	db.First(&p1, product.ID)
	fmt.Printf("Product: %+v\n", p1)

	fmt.Println("\n=== Second Query (From Redis Cache) ===")
	var p2 Product
	db.First(&p2, product.ID)
	fmt.Printf("Product: %+v\n", p2)

	fmt.Println("\n=== Update Product ===")
	db.Model(&product).Update("Price", 899.99)

	fmt.Println("\n=== Third Query (From Database - Cache Cleared) ===")
	var p3 Product
	db.First(&p3, product.ID)
	fmt.Printf("Product: %+v\n", p3)

	// Cleanup
	db.Delete(&product)
	cachePlugin.Close()
}
