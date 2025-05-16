package JobHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (h *Handler) UpdateJob(c *gin.Context) {
	// İş ID'sini al
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Geçersiz iş ilanı ID'si")
		return
	}

	// Kullanıcı ID'sini al
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "Bu işlem için giriş yapmanız gerekiyor")
		return
	}

	// İstek verilerini doğrula
	var input types.JobInput
	if err := utils.ValidateRequest(c, &input); err != nil {
		return
	}

	// İş ilanını güncelle
	job, err := h.JobRepository.UpdateJob(c.Request.Context(), jobID, input, userID.(uuid.UUID))
	if err != nil {
		utils.HandleDatabaseError(c, err, "İş ilanı güncelleme")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "İş ilanı başarıyla güncellendi",
		"data":    job,
	})
}

func (h *Handler) UpdateJobStatus(c *gin.Context) {
	// İş ID'sini al
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Geçersiz iş ilanı ID'si")
		return
	}

	// İstek verilerini doğrula
	var input types.JobStatusInput
	if err := utils.ValidateRequest(c, &input); err != nil {
		return
	}

	// İş ilanı durumunu güncelle
	err = h.JobRepository.UpdateJobStatus(c.Request.Context(), jobID, input.Status)
	if err != nil {
		utils.HandleDatabaseError(c, err, "İş ilanı durumu güncelleme")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "İş ilanı durumu başarıyla güncellendi",
	})
}
