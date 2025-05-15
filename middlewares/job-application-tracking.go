package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-holding/utils"
)

// ApplicationTrackingMiddleware iş başvuru takip oturumunu doğrular
func ApplicationTrackingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Cookie'den token al
		tokenCookie, err := c.Cookie("application_tracking_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized",
				"message": "İş başvurularını görüntülemek için takip kodunu girin",
			})
			c.Abort()
			return
		}

		// Token'ı doğrula ve email al
		email, err := utils.VerifyApplicationTrackingToken(tokenCookie)
		if err != nil {
			// Cookie'yi temizle
			c.SetCookie("application_tracking_token", "", -1, "/", "", false, true)

			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized",
				"message": "Oturum süresi doldu, lütfen tekrar giriş yapın",
			})
			c.Abort()
			return
		}

		// Email'i context'e ekle
		c.Set("tracking_email", email)
		c.Next()
	}
}
