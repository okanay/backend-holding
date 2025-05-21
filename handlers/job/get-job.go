package JobHandler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/services/cache"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (h *Handler) GetJobByID(c *gin.Context) {
	// İş ID'sini al
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Geçersiz iş ilanı ID'si")
		return
	}

	// Cache kontrolü - önbellekte varsa doğrudan dön
	cacheIdentifier := "job:detail:" + jobID.String()
	if h.Cache.TryCache(c, cache.GroupJobs, cacheIdentifier) {
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

	// Yanıt hazırla
	response := gin.H{
		"success": true,
		"data":    job,
	}

	// Yanıtı önbelleğe al
	h.Cache.SaveCache(response, cache.GroupJobs, cacheIdentifier)

	// Cache header'ı ekle
	c.Header("X-Cache", "MISS")

	// Yanıtı döndür
	c.JSON(http.StatusOK, response)
}

// GetJobBySlug bir iş ilanını slug'a göre getirir
func (h *Handler) GetJobBySlug(c *gin.Context) {
	// Slug'ı al
	slug := c.Param("id")
	if slug == "" {
		utils.BadRequest(c, "Geçersiz URL yapısı")
		return
	}

	// Cache kontrolü - önbellekte varsa doğrudan dön
	cacheIdentifier := "job:slug:" + slug
	if h.Cache.TryCache(c, cache.GroupJobs, cacheIdentifier) {
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

	// Yanıt hazırla
	response := gin.H{
		"success": true,
		"data":    job,
	}

	// Yanıtı önbelleğe al
	h.Cache.SaveCache(response, cache.GroupJobs, cacheIdentifier)

	// Cache header'ı ekle
	c.Header("X-Cache", "MISS")

	// Yanıtı döndür
	c.JSON(http.StatusOK, response)
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

	// Cache identifier oluştur - tüm parametreleri içerir
	cacheIdentifier := fmt.Sprintf("job:list:p%d:l%d:s%s:o%s:st%s:c%s:loc%s:q%s",
		page, limit, sortBy, sortOrder, status, category, location, query)

	// Cache kontrolü - önbellekte varsa doğrudan dön
	if h.Cache.TryCache(c, cache.GroupJobs, cacheIdentifier) {
		return
	}

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

	// Yanıt hazırla
	response := gin.H{
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
	}

	// Yanıtı önbelleğe al - liste sorguları için daha kısa TTL kullanıyoruz
	h.Cache.SaveCacheTTL(response, cache.GroupJobs, cacheIdentifier, 5*60*time.Second) // 5 dakika

	// Cache header'ı ekle
	c.Header("X-Cache", "MISS")

	// Yanıtı döndür
	c.JSON(http.StatusOK, response)
}

func (h *Handler) ListPublishedJobs(c *gin.Context) {
	// Sayfalama parametrelerini al
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	// Sıralama ve filtreleme parametrelerini al
	sortBy := c.DefaultQuery("sortBy", "createdAt")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	category := c.DefaultQuery("category", "")
	location := c.DefaultQuery("location", "")
	query := c.DefaultQuery("q", "")

	// Cache identifier oluştur - tüm parametreleri içerir
	cacheIdentifier := fmt.Sprintf("job:published:p%d:l%d:s%s:o%s:c%s:loc%s:q%s",
		page, limit, sortBy, sortOrder, category, location, query)

	// Cache kontrolü - önbellekte varsa doğrudan dön
	if h.Cache.TryCache(c, cache.GroupJobs, cacheIdentifier) {
		return
	}

	// Parametreleri SearchParams yapısına dönüştür
	params := types.JobSearchParams{
		Status:    types.JobStatusPublished,
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

	// Yanıt hazırla
	response := gin.H{
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
	}

	// Yanıtı önbelleğe al - yayınlanmış iş ilanları için daha uzun TTL kullanabiliriz
	h.Cache.SaveCacheTTL(response, cache.GroupJobs, cacheIdentifier, 15*60*time.Second) // 15 dakika

	// Cache header'ı ekle
	c.Header("X-Cache", "MISS")

	// Yanıtı döndür
	c.JSON(http.StatusOK, response)
}
