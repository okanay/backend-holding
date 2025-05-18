package JobHandler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (h *Handler) GetJob(c *gin.Context) {
	// İş ID'sini al
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Geçersiz iş ilanı ID'si")
		return
	}

	// İş ilanını getir
	job, err := h.JobRepository.GetJobByID(c.Request.Context(), jobID)
	if err != nil {
		utils.HandleDatabaseError(c, err, "İş ilanı getirme")
		return
	}

	// İş ilanı bulunamadıysa
	if job.ID == uuid.Nil {
		utils.NotFound(c, "İş ilanı")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    job,
	})
}

// GetJobBySlug bir iş ilanını slug'a göre getirir
func (h *Handler) GetJobBySlug(c *gin.Context) {
	// Slug'ı al
	slug := c.Param("id")
	if slug == "" {
		utils.BadRequest(c, "Geçersiz URL yapısı")
		return
	}

	// İş ilanını getir
	job, err := h.JobRepository.GetJobBySlug(c.Request.Context(), slug)
	if err != nil {
		utils.HandleDatabaseError(c, err, "İş ilanı getirme")
		return
	}

	// İş ilanı bulunamadıysa
	if job.ID == uuid.Nil {
		utils.NotFound(c, "İş ilanı")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    job,
	})
}

// ListJobs tüm iş ilanlarını listeler (filtreleme, sayfalama ve sıralama seçenekleriyle)
func (h *Handler) ListJobs(c *gin.Context) {
	// Sayfalama parametrelerini al
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	// Sıralama ve filtreleme parametrelerini al
	sortBy := c.DefaultQuery("sortBy", "createdAt")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	status := types.JobStatus(c.DefaultQuery("status", ""))
	category := c.DefaultQuery("category", "")
	location := c.DefaultQuery("location", "")
	query := c.DefaultQuery("q", "")

	// Parametreleri SearchParams yapısına dönüştür
	params := types.JobSearchParams{
		Status:    status,
		Category:  category,
		Location:  location,
		Query:     query,
		Page:      page,
		Limit:     limit,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	// İş ilanlarını getir
	jobs, total, err := h.JobRepository.ListJobs(c.Request.Context(), params)
	if err != nil {
		utils.HandleDatabaseError(c, err, "İş ilanları listeleme")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"jobs": jobs,
			"pagination": gin.H{
				"currentPage": page,
				"pageSize":    limit,
				"totalItems":  total,
				"totalPages":  (total + limit - 1) / limit,
			},
		},
	})
}
