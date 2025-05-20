package JobHandler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/utils"
)

func (h *Handler) DeleteJob(c *gin.Context) {
	// İş ID'sini al
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Geçersiz iş ilanı ID'si")
		return
	}

	// Hard delete parametresini kontrol et (varsayılan: false)
	hardDelete, _ := strconv.ParseBool(c.DefaultQuery("hard", "false"))

	// İş ilanını sil
	err = h.JobRepository.DeleteJob(c.Request.Context(), jobID, hardDelete)
	if err != nil {
		utils.HandleDatabaseError(c, err, "İş ilanı silme")
		return
	}

	message := "İş ilanı başarıyla silindi"
	if hardDelete {
		message = "İş ilanı tamamen silindi"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}
