package ContentHandler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/utils"
)

func (h *Handler) DeleteContent(c *gin.Context) {
	// ID parse
	contentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Geçersiz içerik ID'si")
		return
	}

	// Hard delete kontrolü
	hardDelete, _ := strconv.ParseBool(c.DefaultQuery("hard", "false"))

	// Silme işlemi
	if hardDelete {
		err = h.Repository.HardDeleteContent(c.Request.Context(), contentID)
	} else {
		err = h.Repository.SoftDeleteContent(c.Request.Context(), contentID)
	}

	if err != nil {
		utils.HandleDatabaseError(c, err, "İçerik silme")
		return
	}

	// Cache temizle
	h.Cache.ClearGroup(Group)

	// Response mesajı
	message := "İçerik geçici olarak silindi"
	if hardDelete {
		message = "İçerik kalıcı olarak silindi"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}

func (h *Handler) RestoreContent(c *gin.Context) {
	// ID parse
	contentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Geçersiz içerik ID'si")
		return
	}

	// Yeni slug al
	var input struct {
		Slug string `json:"slug" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Yeni slug belirtmelisiniz")
		return
	}

	// Geri yükle
	err = h.Repository.RestoreContent(c.Request.Context(), contentID, input.Slug)
	if err != nil {
		utils.HandleDatabaseError(c, err, "İçerik geri yükleme")
		return
	}

	// Cache temizle
	h.Cache.ClearGroup(Group)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "İçerik geri yüklendi",
	})
}
