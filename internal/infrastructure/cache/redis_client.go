package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClient wraps the Redis client with common functionality
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisClient creates a new Redis client wrapper
func NewRedisClient(client *redis.Client) *RedisClient {
	return &RedisClient{
		client: client,
		ctx:    context.Background(),
	}
}

// GetClient returns the underlying Redis client
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// GetContext returns the context
func (r *RedisClient) GetContext() context.Context {
	return r.ctx
}

// GenerateKey creates a namespaced key
func (r *RedisClient) GenerateKey(prefix, key string) string {
	return fmt.Sprintf("%s:%s", prefix, key)
}

// Set stores data with TTL
func (r *RedisClient) Set(key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(r.ctx, key, value, ttl).Err()
}

// Get retrieves data
func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

// Delete removes data
func (r *RedisClient) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// Exists checks if key exists
func (r *RedisClient) Exists(key string) (bool, error) {
	count, err := r.client.Exists(r.ctx, key).Result()
	return count > 0, err
}

// SetBytes stores byte data with TTL
func (r *RedisClient) SetBytes(key string, value []byte, ttl time.Duration) error {
	return r.client.Set(r.ctx, key, value, ttl).Err()
}

// GetBytes retrieves byte data (backwards-compatible wrapper)
func (r *RedisClient) GetBytes(key string) ([]byte, error) {
	return r.GetBytesWithContext(r.ctx, key)
}

// SetWithContext stores data with TTL using provided context
func (r *RedisClient) SetWithContext(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// GetWithContext retrieves data using provided context
func (r *RedisClient) GetWithContext(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// DeleteWithContext removes data using provided context
func (r *RedisClient) DeleteWithContext(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// ExistsWithContext checks if key exists using provided context
func (r *RedisClient) ExistsWithContext(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

// SetBytesWithContext stores byte data with TTL using provided context
func (r *RedisClient) SetBytesWithContext(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// GetBytesWithContext retrieves byte data using provided context
func (r *RedisClient) GetBytesWithContext(ctx context.Context, key string) ([]byte, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}