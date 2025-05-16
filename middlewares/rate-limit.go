package middlewares

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func RateLimiterMiddleware(limit int, period time.Duration) gin.HandlerFunc {
	// Her IP için ayrı bir limiter oluştur
	limiters := make(map[string]*rate.Limiter)
	var mu sync.Mutex

	return func(c *gin.Context) {
		// İstemci IP'sini al
		ip := c.ClientIP()

		// IP için limiter'ı bul veya oluştur
		mu.Lock()
		limiter, exists := limiters[ip]
		if !exists {
			// Yeni limiter oluştur (rate.Every ile periyot başına istek sayısını belirle)
			limiter = rate.NewLimiter(rate.Every(period/time.Duration(limit)), limit)
			limiters[ip] = limiter
		}
		mu.Unlock()

		// Bu istek için izin kontrol et
		if !limiter.Allow() {
			// Rate limit aşıldı
			c.JSON(429, gin.H{
				"success": false,
				"error":   "rate_limit_exceeded",
				"message": "Çok fazla istek gönderdiniz. Lütfen daha sonra tekrar deneyin.",
			})
			c.Abort()
			return
		}

		// Sonraki middleware'e geç
		c.Next()
	}
}
