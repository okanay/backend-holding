package ContentHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/services/cache"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils" //
)

func (h *Handler) UpdateContent(c *gin.Context) {
	contentIDStr := c.Param("id")
	contentID, err := uuid.Parse(contentIDStr)
	if err != nil {
		utils.BadRequest(c, "Geçersiz içerik ID'si") //
		return
	}

	userIDValue, exists := c.Get("user_id") //
	if !exists {
		utils.Unauthorized(c, "Bu işlem için giriş yapmanız gerekiyor") //
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		utils.SendError(c, "internal_error", "Kullanıcı ID'si formatı geçersiz.")
		return
	}

	var input types.ContentInput
	if err := utils.ValidateRequest(c, &input); err != nil { //
		return
	}

	updatedContent, err := h.Repository.UpdateContent(c.Request.Context(), contentID, input, userID)
	if err != nil {
		utils.HandleDatabaseError(c, err, "Basın içeriği güncelleme") //
		return
	}

	h.Cache.ClearGroup(Group)    //
	c.JSON(http.StatusOK, gin.H{ //
		"success": true,
		"message": "Basın içeriği başarıyla güncellendi",
		"data":    mapContentToView(updatedContent),
	})
}

func (h *Handler) UpdateContentStatus(c *gin.Context) {
	contentIDStr := c.Param("id") //
	contentID, err := uuid.Parse(contentIDStr)
	if err != nil {
		utils.BadRequest(c, "Geçersiz içerik ID'si") //
		return
	}

	var input types.ContentStatusInput
	if err := utils.ValidateRequest(c, &input); err != nil { //
		return
	}

	err = h.Repository.UpdateContentStatus(c.Request.Context(), contentID, input.Status)
	if err != nil {
		utils.HandleDatabaseError(c, err, "Basın içeriği durumu güncelleme") //
		return
	}

	h.Cache.ClearGroup(cache.GroupContent)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Basın içeriği durumu başarıyla güncellendi",
	})
}
