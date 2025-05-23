package ContentHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/services/cache"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils" //
)

func (h *Handler) CreateContent(c *gin.Context) {
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

	createdContent, err := h.Repository.CreateContent(c.Request.Context(), input, userID)
	if err != nil {
		utils.HandleDatabaseError(c, err, "Basın içeriği oluşturma") //
		return
	}

	h.Cache.ClearGroup(cache.GroupContent)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Basın içeriği başarıyla oluşturuldu",
		"data":    mapContentToView(createdContent),
	})
}
