package JobHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/services/cache"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (h *Handler) CreateJob(c *gin.Context) {
	// Kullanıcı ID'sini context'ten al
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

	// İş ilanını oluştur
	job, err := h.JobRepository.CreateJob(c.Request.Context(), input, userID.(uuid.UUID))
	if err != nil {
		utils.HandleDatabaseError(c, err, "İş ilanı oluşturma")
		return
	}

	h.Cache.ClearGroup(cache.GroupJobs)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "İş ilanı başarıyla oluşturuldu",
		"data":    job,
	})
}
