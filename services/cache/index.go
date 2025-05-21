// cache/cache.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8" // veya github.com/redis/go-redis/v8
)

// Cache grupları - gerektiğinde ekleyebilirsiniz
const (
	GroupJobs = "jobs"
)

// CacheService, tüm cache implementasyonları için ortak arayüz
type CacheService interface {
	TryCache(ctx *gin.Context, group, identifier string) bool
	SaveCache(response any, group, identifier string) error
	SaveCacheTTL(response any, group, identifier string, ttl time.Duration) error
	ClearGroup(group string)
	ClearAll()
	Stop()
}

// NewCacheService, ortam değişkenlerine göre uygun cache servisini döndürür
func NewCacheService(defaultTTL time.Duration) CacheService {
	// Ortam değişkenlerinden Redis yapılandırmasını al
	useRedis := os.Getenv("USE_REDIS")

	// Redis kullanılacak mı?
	if useRedis == "true" {
		redisAddr := os.Getenv("REDIS_ADDR")
		if redisAddr == "" {
			redisAddr = "localhost:6379" // Varsayılan adres
		}

		redisPassword := os.Getenv("REDIS_PASSWORD")
		redisDB := 0 // Varsayılan DB

		// Redis bağlantısını oluştur ve cache servisini döndür
		return NewRedisCache(redisAddr, redisPassword, redisDB, defaultTTL)
	}

	// Redis devre dışı veya yapılandırılmamışsa in-memory cache kullan
	return NewInMemoryCache(defaultTTL)
}

// ===== IN-MEMORY CACHE IMPLEMENTATION =====

// InMemoryCache in-memory önbellekleme için yapı
type InMemoryCache struct {
	mu              sync.RWMutex
	data            map[string]cacheItem
	ttl             time.Duration
	stopCleanup     chan struct{}
	cleanupInterval time.Duration
}

// cacheItem önbellekteki bir veriyi ve metadata'sını temsil eder
type cacheItem struct {
	value    []byte
	cachedAt time.Time
	ttl      time.Duration // Opsiyonel TTL
}

// NewInMemoryCache yeni bir in-memory cache instance'ı oluşturur
func NewInMemoryCache(ttl time.Duration) *InMemoryCache {
	cache := &InMemoryCache{
		data:            make(map[string]cacheItem),
		ttl:             ttl,
		stopCleanup:     make(chan struct{}),
		cleanupInterval: 30 * time.Minute,
	}

	// Periyodik temizleme başlat
	go cache.startCleanupRoutine()
	return cache
}

// TryCache önbellekteki veriyi kontrol eder ve varsa yanıt olarak döndürür
func (c *InMemoryCache) TryCache(ctx *gin.Context, group, identifier string) bool {
	cacheKey := fmt.Sprintf("%s:%s", group, identifier)

	c.mu.RLock()
	item, exists := c.data[cacheKey]
	c.mu.RUnlock()

	if !exists {
		return false
	}

	// TTL kontrolü
	now := time.Now()

	// Özel TTL kontrolü
	if item.ttl > 0 && now.Sub(item.cachedAt) > item.ttl {
		c.mu.Lock()
		delete(c.data, cacheKey)
		c.mu.Unlock()
		return false
	}

	// Genel TTL kontrolü
	if now.Sub(item.cachedAt) > c.ttl {
		c.mu.Lock()
		delete(c.data, cacheKey)
		c.mu.Unlock()
		return false
	}

	// Cache hit - önbellekteki veriyi dön
	ctx.Data(http.StatusOK, "application/json", item.value)
	ctx.Header("X-Cache", "HIT")
	return true
}

// SaveCache yanıtı önbelleğe alır
func (c *InMemoryCache) SaveCache(response any, group, identifier string) error {
	jsonData, err := json.Marshal(response)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("%s:%s", group, identifier)

	c.mu.Lock()
	c.data[cacheKey] = cacheItem{
		value:    jsonData,
		cachedAt: time.Now(),
	}
	c.mu.Unlock()

	return nil
}

// SaveCacheTTL özel TTL ile yanıtı önbelleğe alır
func (c *InMemoryCache) SaveCacheTTL(response any, group, identifier string, ttl time.Duration) error {
	jsonData, err := json.Marshal(response)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("%s:%s", group, identifier)

	c.mu.Lock()
	c.data[cacheKey] = cacheItem{
		value:    jsonData,
		cachedAt: time.Now(),
		ttl:      ttl,
	}
	c.mu.Unlock()

	return nil
}

// ClearGroup bir grubu önbellekten temizler
func (c *InMemoryCache) ClearGroup(group string) {
	prefix := group + ":"

	c.mu.Lock()
	defer c.mu.Unlock()

	// Belirli bir önekle başlayan tüm anahtarları temizle
	for key := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.data, key)
		}
	}
}

// ClearAll tüm önbelleği temizler
func (c *InMemoryCache) ClearAll() {
	c.mu.Lock()
	c.data = make(map[string]cacheItem)
	c.mu.Unlock()
}

// Stop temizleme goroutine'ini durdurur
func (c *InMemoryCache) Stop() {
	close(c.stopCleanup)
}

// startCleanupRoutine periyodik temizleme rutini
func (c *InMemoryCache) startCleanupRoutine() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanExpiredItems()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanExpiredItems süresi dolmuş cache öğelerini temizler
func (c *InMemoryCache) cleanExpiredItems() {
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.data {
		// Özel TTL kontrolü
		if item.ttl > 0 && now.Sub(item.cachedAt) > item.ttl {
			delete(c.data, key)
			continue
		}

		// Genel TTL kontrolü
		if now.Sub(item.cachedAt) > c.ttl {
			delete(c.data, key)
		}
	}
}

// ===== REDIS CACHE IMPLEMENTATION =====

// RedisCache Redis tabanlı önbellekleme için yapı
type RedisCache struct {
	client     *redis.Client
	ctx        context.Context
	defaultTTL time.Duration
}

// NewRedisCache yeni bir Redis cache instance'ı oluşturur
func NewRedisCache(addr string, password string, db int, ttl time.Duration) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{
		client:     client,
		ctx:        context.Background(),
		defaultTTL: ttl,
	}
}

// TryCache önbellekteki veriyi kontrol eder ve varsa yanıt olarak döndürür
func (c *RedisCache) TryCache(ctx *gin.Context, group, identifier string) bool {
	cacheKey := fmt.Sprintf("%s:%s", group, identifier)

	// Redis'ten veriyi al
	val, err := c.client.Get(c.ctx, cacheKey).Bytes()
	if err != nil {
		// Redis'te yok veya bir hata oluştu
		return false
	}

	// Cache hit - önbellekteki veriyi dön
	ctx.Data(http.StatusOK, "application/json", val)
	ctx.Header("X-Cache", "HIT")
	return true
}

// SaveCache yanıtı önbelleğe alır
func (c *RedisCache) SaveCache(response any, group, identifier string) error {
	jsonData, err := json.Marshal(response)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("%s:%s", group, identifier)

	// Redis'e kaydet
	return c.client.Set(c.ctx, cacheKey, jsonData, c.defaultTTL).Err()
}

// SaveCacheTTL özel TTL ile yanıtı önbelleğe alır
func (c *RedisCache) SaveCacheTTL(response any, group, identifier string, ttl time.Duration) error {
	jsonData, err := json.Marshal(response)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("%s:%s", group, identifier)

	// Redis'e özel TTL ile kaydet
	return c.client.Set(c.ctx, cacheKey, jsonData, ttl).Err()
}

// ClearGroup bir grubu önbellekten temizler
func (c *RedisCache) ClearGroup(group string) {
	prefix := group + ":"

	// Redis'te desen araması yap
	iter := c.client.Scan(c.ctx, 0, prefix+"*", 0).Iterator()

	// Bulunan tüm anahtarları sil
	for iter.Next(c.ctx) {
		c.client.Del(c.ctx, iter.Val())
	}
}

// ClearAll tüm önbelleği temizler
func (c *RedisCache) ClearAll() {
	c.client.FlushAll(c.ctx)
}

// Stop Redis bağlantısını kapatır
func (c *RedisCache) Stop() {
	c.client.Close()
}
