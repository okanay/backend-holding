package JobHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (h *Handler) CreateJobApplication(c *gin.Context) {
	// İş ID'sini al
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Geçersiz iş ilanı ID'si")
		return
	}

	// İstek verilerini doğrula
	var input types.JobApplicationInput
	if err := utils.ValidateRequest(c, &input); err != nil {
		return
	}

	// Başvuruyu oluştur
	application, err := h.JobRepository.CreateJobApplication(c.Request.Context(), jobID, input)
	if err != nil {
		utils.HandleDatabaseError(c, err, "Başvuru oluşturma")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Başvurunuz başarıyla alındı",
		"data":    application,
	})
}
