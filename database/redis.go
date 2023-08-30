package database

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

// Redis contains the Redis client instance
type Redis struct {
	client *redis.Client
}

// RedisImpl defines the interface for interacting with Redis
type RedisImpl interface {
	Get(key string) *redis.StringCmd
	Set(key string, value interface{}) *redis.StatusCmd
	Del(keys ...string) *redis.IntCmd
	Close() error
}

// Get retrieves the value associated with the given key from Redis
func (r *Redis) Get(key string) *redis.StringCmd {
	return r.client.Get(context.Background(), key)
}

// Set sets the value associated with the given key in Redis
func (r *Redis) Set(key string, value interface{}) *redis.StatusCmd {
	return r.client.Set(context.Background(), key, value, 0)
}

// Del deletes the keys from Redis
func (r *Redis) Del(keys ...string) *redis.IntCmd {
	return r.client.Del(context.Background(), keys...)
}

// Close closes the Redis client connection
func (r *Redis) Close() error {
	return r.client.Close()
}

// Setup connects to the Redis database and returns a Redis instance
func SetupRedis(addr string, password string, db int) (RedisImpl, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("error while connecting to Redis: %v", err)
	}
	return &Redis{client: client}, nil
}
