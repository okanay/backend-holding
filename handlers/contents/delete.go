package ContentHandler

import (
	"net/http"
	"strconv" //

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/services/cache"
	"github.com/okanay/backend-holding/utils" //
)

func (h *Handler) DeleteContent(c *gin.Context) {
	contentIDStr := c.Param("id")
	contentID, err := uuid.Parse(contentIDStr)
	if err != nil {
		utils.BadRequest(c, "Geçersiz içerik ID'si") //
		return
	}

	hardDelete, _ := strconv.ParseBool(c.DefaultQuery("hard", "false")) //

	if hardDelete {
		err = h.Repository.HardDeleteContent(c.Request.Context(), contentID)
	} else {
		err = h.Repository.SoftDeleteContent(c.Request.Context(), contentID)
	}

	if err != nil {
		utils.HandleDatabaseError(c, err, "Basın içeriği silme") //
		return
	}

	message := "Basın içeriği başarıyla geçici olarak silindi (soft delete)"
	if hardDelete {
		message = "Basın içeriği kalıcı olarak silindi (hard delete)"
	}

	h.Cache.ClearGroup(cache.GroupContent)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}
