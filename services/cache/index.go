// cache/cache.go
package cache

import (
	"sync"
	"time"
)

// ---- Veri Yapıları ----

// Cache genel amaçlı bir önbellek yapısı
type Cache struct {
	mu              sync.RWMutex
	data            map[string]cacheItem
	ttl             time.Duration
	stopCleanup     chan struct{}
	lastCleanupTime time.Time     // Son temizleme zamanı
	cleanupInterval time.Duration // Temizleme aralığı
	startTime       time.Time     // Cache başlatma zamanı
}

// cacheItem önbellekteki bir veriyi ve metadata'sını temsil eder
type cacheItem struct {
	value    []byte
	cachedAt time.Time
	ttl      time.Duration // Opsiyonel TTL
}

// ---- Temel İşlemler ----

// NewCache yeni bir önbellek oluşturur
func NewCache(ttl time.Duration) *Cache {
	cache := &Cache{
		data:            make(map[string]cacheItem),
		ttl:             ttl,
		stopCleanup:     make(chan struct{}),
		cleanupInterval: 30 * time.Minute, // Varsayılan temizleme aralığı
		startTime:       time.Now(),
	}

	// Otomatik temizleme goroutine'ini başlat
	go cache.startCleanupRoutine()

	return cache
}

// Set verilen anahtarla bir değeri önbelleğe kaydeder
func (c *Cache) Set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheItem{
		value:    value,
		cachedAt: time.Now(),
	}
}

// SetWithTTL özel TTL ile bir değeri önbelleğe kaydeder
func (c *Cache) SetWithTTL(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheItem{
		value:    value,
		cachedAt: time.Now(),
		ttl:      ttl,
	}
}

// Get bir anahtara karşılık gelen değeri önbellekten döndürür
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	item, exists := c.data[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Özel TTL kontrolü
	if item.ttl > 0 && time.Since(item.cachedAt) > item.ttl {
		// TTL dolmuş, veriyi sil ve false dön
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		return nil, false
	}

	// Genel TTL kontrolü
	if time.Since(item.cachedAt) > c.ttl {
		return nil, false
	}

	return item.value, true
}

// GetAllWithPrefix belirli bir önekle başlayan tüm anahtarları ve değerlerini döndürür
func (c *Cache) GetAllWithPrefix(prefix string) map[string][]byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string][]byte)
	for key, item := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			// TTL kontrolü
			if item.ttl > 0 && time.Since(item.cachedAt) > item.ttl {
				continue // Süresi dolmuş
			}

			if time.Since(item.cachedAt) > c.ttl {
				continue // Genel TTL süresi dolmuş
			}

			result[key] = item.value
		}
	}

	return result
}

// Delete bir anahtarı önbellekten siler
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

// Clear tüm önbelleği temizler
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]cacheItem)
}

func (c *Cache) ClearAIRateLimits() {
	c.mu.Lock()
	defer c.mu.Unlock()

	prefix := "ai_rate_limit:"
	for key := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.data, key)
		}
	}
}

// ClearExceptPrefixes belirli öneklerle başlayan anahtarlar dışındaki tüm anahtarları temizler
func (c *Cache) ClearExceptPrefixes(prefixes []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keysToDelete := make([]string, 0)
	for key := range c.data {
		keep := false
		for _, prefix := range prefixes {
			if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
				keep = true
				break
			}
		}
		if !keep {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(c.data, key)
	}
}

// ClearPrefix belirli bir önekle başlayan tüm anahtarları temizler
func (c *Cache) ClearPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.data, key)
		}
	}
}

// ---- Timer İşlemleri ----

// startCleanupRoutine temizleme rutinini başlatır
func (c *Cache) startCleanupRoutine() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			c.lastCleanupTime = time.Now()
			c.mu.Unlock()
			c.cleanupExpiredItems()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanupExpiredItems süresi dolmuş cache item'larını temizler
func (c *Cache) cleanupExpiredItems() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
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

// Stop cleanup goroutine'ini durdurur (graceful shutdown için)
func (c *Cache) Stop() {
	close(c.stopCleanup)
}
