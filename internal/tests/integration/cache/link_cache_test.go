package infrastructure_test

import (
	"testing"
	"time"

	"service-short-link/internal/domain"
	icache "service-short-link/internal/infrastructure/cache"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func newRedisClient(t *testing.T) (*redis.Client, func()) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	return client, func() { client.Close(); s.Close() }
}

func TestLinkCache_SetGet_Link(t *testing.T) {
	client, cleanup := newRedisClient(t)
	defer cleanup()

	cache := icache.NewLinkCache(client)
	ttl := 30 * time.Second
	now := time.Now().Add(ttl)
	link := &domain.CachedLink{ID: 1, OriginalURL: "https://example.com", IsActive: true, ExpiresAt: &now}

	err := cache.Set("abcde", link, ttl)
	assert.NoError(t, err)

	got, err := cache.Get("abcde")
	assert.NoError(t, err)
	assert.Equal(t, link.OriginalURL, got.OriginalURL)
}

func TestLinkCache_QRCode_SetGet(t *testing.T) {
	client, cleanup := newRedisClient(t)
	defer cleanup()

	cache := icache.NewLinkCache(client)
	data := []byte{1, 2, 3}
	err := cache.SetQRCode("abcde", data, 10*time.Second)
	assert.NoError(t, err)

	got, err := cache.GetQRCode("abcde")
	assert.NoError(t, err)
	assert.Equal(t, data, got)
}

func TestLinkCache_Get_MissingReturnsDomainError(t *testing.T) {
	client, cleanup := newRedisClient(t)
	defer cleanup()

	cache := icache.NewLinkCache(client)
	got, err := cache.Get("notfound")
	assert.Nil(t, got)
	assert.Equal(t, domain.ErrCacheNotFound, err)
}

func TestLinkCache_ConnectionError(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	cache := icache.NewLinkCache(client)
	s.Close()

	e := cache.Set("k", &domain.CachedLink{ID: 1, OriginalURL: "u", IsActive: true}, 1*time.Second)
	assert.Error(t, e)

	_, ge := cache.Get("k")
	assert.Error(t, ge)

	be := cache.SetQRCode("k", []byte{1}, 1*time.Second)
	assert.Error(t, be)
}

func TestLinkCache_TTL_Expiry(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer s.Close()

	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	cache := icache.NewLinkCache(client)

	ttl := 5 * time.Second
	exp := time.Now().Add(ttl)
	err = cache.Set("ttlkey", &domain.CachedLink{ID: 1, OriginalURL: "https://ttl.example", IsActive: true, ExpiresAt: &exp}, ttl)
	assert.NoError(t, err)

	_, err = cache.Get("ttlkey")
	assert.NoError(t, err)

	s.FastForward(10 * time.Second)

	_, err = cache.Get("ttlkey")
	assert.Equal(t, domain.ErrCacheNotFound, err)
}

func TestLinkCache_KeyNamespaces_LinkAndQR(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer s.Close()

	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	cache := icache.NewLinkCache(client)

	exp := time.Now().Add(1 * time.Minute)
	err = cache.Set("abcde", &domain.CachedLink{ID: 1, OriginalURL: "u", IsActive: true, ExpiresAt: &exp}, 10*time.Second)
	assert.NoError(t, err)

	err = cache.SetQRCode("abcde", []byte{7, 7}, 10*time.Second)
	assert.NoError(t, err)

	assert.True(t, s.Exists("link:abcde"))
	assert.True(t, s.Exists("qr:abcde"))
}
