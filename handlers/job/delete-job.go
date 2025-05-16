package JobHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (h *Handler) DeleteJob(c *gin.Context) {
	// İş ID'sini al
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Geçersiz iş ilanı ID'si")
		return
	}

	// İlanın durumunu "deleted" olarak güncelle
	err = h.JobRepository.UpdateJobStatus(c.Request.Context(), jobID, types.JobStatusDeleted)
	if err != nil {
		utils.HandleDatabaseError(c, err, "İş ilanı silme")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "İş ilanı başarıyla silindi",
	})
}
