package ContentHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (h *Handler) CreateContent(c *gin.Context) {
	// Kullanıcı kontrolü
	userIDValue, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "Giriş yapmanız gerekiyor")
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		utils.InternalError(c, "Kullanıcı bilgisi alınamadı")
		return
	}

	// Input validasyonu
	var input types.ContentInput
	if err := utils.ValidateRequest(c, &input); err != nil {
		return
	}

	// İçerik oluştur
	content, err := h.Repository.CreateContent(c.Request.Context(), input, userID)
	if err != nil {
		utils.HandleDatabaseError(c, err, "İçerik oluşturma")
		return
	}

	// Cache temizle
	h.Cache.ClearGroup(Group)

	// Response
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "İçerik başarıyla oluşturuldu",
		"data":    mapContentToView(content),
	})
}
