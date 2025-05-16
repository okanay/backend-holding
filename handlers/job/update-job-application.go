package JobHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (h *Handler) UpdateJobApplicationStatus(c *gin.Context) {
	// Başvuru ID'sini al
	applicationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Geçersiz başvuru ID'si")
		return
	}

	// İstek verilerini doğrula
	var input types.JobApplicationStatusInput
	if err := utils.ValidateRequest(c, &input); err != nil {
		return
	}

	// Başvuru durumunu güncelle
	err = h.JobRepository.UpdateJobApplicationStatus(c.Request.Context(), applicationID, input.Status)
	if err != nil {
		utils.HandleDatabaseError(c, err, "Başvuru durumu güncelleme")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Başvuru durumu başarıyla güncellendi",
	})
}
