package gormcache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisAdapter is a Redis cache implementation
type RedisAdapter struct {
	client *redis.Client
}

// RedisAdapterConfig holds configuration for Redis adapter
type RedisAdapterConfig struct {
	Addr     string // Redis server address (default: "localhost:6379")
	Password string // Redis password (default: "")
	DB       int    // Redis database (default: 0)
}

// NewRedisAdapter creates a new Redis cache adapter
func NewRedisAdapter(config RedisAdapterConfig) *RedisAdapter {
	if config.Addr == "" {
		config.Addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	return &RedisAdapter{
		client: client,
	}
}

// NewRedisAdapterWithClient creates a new Redis adapter with existing client
func NewRedisAdapterWithClient(client *redis.Client) *RedisAdapter {
	return &RedisAdapter{
		client: client,
	}
}

// Get retrieves a value from Redis cache
func (r *RedisAdapter) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, err
	}
	return val, err
}

// Set stores a value in Redis cache
func (r *RedisAdapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a value from Redis cache
func (r *RedisAdapter) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// DeletePattern removes all keys matching the pattern
func (r *RedisAdapter) DeletePattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()

	pipe := r.client.Pipeline()
	for iter.Next(ctx) {
		pipe.Del(ctx, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Clear removes all cached data in the current database
func (r *RedisAdapter) Clear(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// Close closes the Redis connection
func (r *RedisAdapter) Close() error {
	return r.client.Close()
}
