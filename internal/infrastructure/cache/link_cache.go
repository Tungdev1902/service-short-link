package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"service-short-link/internal/domain"
	"service-short-link/pkg/logger"

	"github.com/go-redis/redis/v8"
)

const (
	linkKeyPrefix = "link"
	qrKeyPrefix   = "qr"
)

type linkCache struct {
	redisClient *RedisClient
}

// NewLinkCache creates a new Link cache instance
func NewLinkCache(client *redis.Client) domain.LinkCache {
	return &linkCache{
		redisClient: NewRedisClient(client),
	}
}

// Set stores a cached link in the cache with TTL
func (c *linkCache) Set(key string, link *domain.CachedLink, ttl time.Duration) error {
	data, err := json.Marshal(link)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "LinkCache.Set: failed to marshal cached link", "key="+key, "error_type=json_marshal_failed")
		return fmt.Errorf("failed to marshal cached link: %w", err)
	}

	cacheKey := c.redisClient.GenerateKey(linkKeyPrefix, key)
	err = c.redisClient.Set(cacheKey, data, ttl)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "LinkCache.Set: failed to set link cache", "key="+key, "cache_key="+cacheKey, "error_type=redis_set_failed")
		return fmt.Errorf("failed to set link cache: %w", err)
	}

	return nil
}

// Get retrieves a cached link from the cache
func (c *linkCache) Get(key string) (*domain.CachedLink, error) {
	cacheKey := c.redisClient.GenerateKey(linkKeyPrefix, key)
	data, err := c.redisClient.Get(cacheKey)
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrCacheNotFound
		}
		logger.ErrorWithCockroachSimple(err, "LinkCache.Get: failed to get link cache", "key="+key, "cache_key="+cacheKey, "error_type=redis_get_failed")
		return nil, fmt.Errorf("failed to get link cache: %w", err)
	}

	var cachedLink domain.CachedLink
	err = json.Unmarshal([]byte(data), &cachedLink)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "LinkCache.Get: failed to unmarshal cached link", "key="+key, "cache_key="+cacheKey, "error_type=json_unmarshal_failed")
		return nil, fmt.Errorf("failed to unmarshal cached link: %w", err)
	}

	return &cachedLink, nil
}

func (c *linkCache) SetQRCode(shortCode string, qrData []byte, ttl time.Duration) error {
	cacheKey := c.redisClient.GenerateKey(qrKeyPrefix, shortCode)
	err := c.redisClient.SetBytes(cacheKey, qrData, ttl)
	if err != nil {
		logger.ErrorWithCockroachSimple(err, "LinkCache.SetQRCode: failed to set QR code cache", "short_code="+shortCode, "cache_key="+cacheKey, "error_type=redis_set_bytes_failed")
		return fmt.Errorf("failed to set QR code cache: %w", err)
	}

	return nil
}

// GetQRCode retrieves QR code data from the cache
func (c *linkCache) GetQRCode(shortCode string) ([]byte, error) {
	cacheKey := c.redisClient.GenerateKey(qrKeyPrefix, shortCode)
	data, err := c.redisClient.GetBytes(cacheKey)
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrCacheNotFound
		}
		logger.ErrorWithCockroachSimple(err, "LinkCache.GetQRCode: failed to get QR code cache", "short_code="+shortCode, "cache_key="+cacheKey, "error_type=redis_get_bytes_failed")
		return nil, fmt.Errorf("failed to get QR code cache: %w", err)
	}

	return data, nil
}
