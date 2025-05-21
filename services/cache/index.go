// cache/cache.go
package cache

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Cache grupları - gerektiğinde ekleyebilirsiniz
const (
	GroupJobs = "jobs"
)

// Cache in-memory önbellekleme için basit bir yapı
type Cache struct {
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

// NewCache yeni bir cache instance'ı oluşturur
func NewCache(ttl time.Duration) *Cache {
	cache := &Cache{
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
// Eğer önbellekte yoksa false döner ve normal işlem sürdürülür
func (c *Cache) TryCache(ctx *gin.Context, group, identifier string) bool {
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
func (c *Cache) SaveCache(response any, group, identifier string) error {
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
func (c *Cache) SaveCacheTTL(response any, group, identifier string, ttl time.Duration) error {
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
func (c *Cache) ClearGroup(group string) {
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
func (c *Cache) ClearAll() {
	c.mu.Lock()
	c.data = make(map[string]cacheItem)
	c.mu.Unlock()
}

// Stop temizleme goroutine'ini durdurur
func (c *Cache) Stop() {
	close(c.stopCleanup)
}

// startCleanupRoutine periyodik temizleme rutini
func (c *Cache) startCleanupRoutine() {
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
func (c *Cache) cleanExpiredItems() {
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
