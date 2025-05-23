package ContentHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (h *Handler) UpdateContent(c *gin.Context) {
	// ID parse
	contentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Geçersiz içerik ID'si")
		return
	}

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

	// Güncelle
	content, err := h.Repository.UpdateContent(c.Request.Context(), contentID, input, userID)
	if err != nil {
		utils.HandleDatabaseError(c, err, "İçerik güncelleme")
		return
	}

	// Cache temizle
	h.Cache.ClearGroup(Group)

	// Response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "İçerik başarıyla güncellendi",
		"data":    mapContentToView(content),
	})
}

func (h *Handler) UpdateContentStatus(c *gin.Context) {
	// ID parse
	contentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Geçersiz içerik ID'si")
		return
	}

	// Input validasyonu
	var input types.ContentStatusInput
	if err := utils.ValidateRequest(c, &input); err != nil {
		return
	}

	// Status güncelle
	err = h.Repository.UpdateContentStatus(c.Request.Context(), contentID, input.Status)
	if err != nil {
		utils.HandleDatabaseError(c, err, "Durum güncelleme")
		return
	}

	// Cache temizle
	h.Cache.ClearGroup(Group)

	// Response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "İçerik durumu güncellendi",
	})
}
