package JobHandler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (h *Handler) ListJobApplications(c *gin.Context) {
	// Sayfalama parametrelerini al
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Sıralama ve filtreleme parametrelerini al
	sortBy := c.DefaultQuery("sortBy", "createdAt")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	status := c.DefaultQuery("status", "")
	email := c.DefaultQuery("email", "")
	startDate := c.DefaultQuery("startDate", "")
	endDate := c.DefaultQuery("endDate", "")

	// İş ID'sini al (opsiyonel)
	var jobID uuid.UUID
	jobIDStr := c.DefaultQuery("jobId", "")
	if jobIDStr != "" {
		var err error
		jobID, err = uuid.Parse(jobIDStr)
		if err != nil {
			utils.BadRequest(c, "Geçersiz iş ilanı ID'si")
			return
		}
	}

	// Parametreleri SearchParams yapısına dönüştür
	params := types.JobApplicationSearchParams{
		JobID:     jobID,
		Status:    status,
		Email:     email,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		Limit:     limit,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	// Başvuruları getir
	applications, total, err := h.JobRepository.ListJobsApplications(c.Request.Context(), params)
	if err != nil {
		utils.HandleDatabaseError(c, err, "Başvurular listeleme")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"applications": applications,
			"pagination": gin.H{
				"currentPage": page,
				"pageSize":    limit,
				"totalItems":  total,
				"totalPages":  (total + limit - 1) / limit,
			},
		},
	})
}

func (h *Handler) GetJobApplication(c *gin.Context) {
	// Başvuru ID'sini al
	applicationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Geçersiz başvuru ID'si")
		return
	}

	// Başvuruyu getir
	application, err := h.JobRepository.GetJobApplicationByID(c.Request.Context(), applicationID)
	if err != nil {
		utils.HandleDatabaseError(c, err, "Başvuru getirme")
		return
	}

	// Başvuru bulunamadıysa
	if application.ID == uuid.Nil {
		utils.NotFound(c, "Başvuru")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    application,
	})
}

func (h *Handler) GetJobApplicationsByEmail(c *gin.Context) {
	// Middleware'den "tracking_email" değerini al
	emailAny, exists := c.Get("tracking_email")
	if !exists {
		utils.Unauthorized(c, "Oturum bilgisi bulunamadı")
		return
	}

	email, ok := emailAny.(string)
	if !ok || email == "" {
		utils.Unauthorized(c, "Geçersiz oturum bilgisi")
		return
	}

	// Başvuruları getir
	applications, err := h.JobRepository.GetJobApplicationsByEmail(c.Request.Context(), email)
	if err != nil {
		utils.HandleDatabaseError(c, err, "Başvurular getirme")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    applications,
	})
}
