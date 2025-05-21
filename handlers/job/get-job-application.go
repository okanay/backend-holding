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

func (h *Handler) ListJobApplications(c *gin.Context) {
	// Sayfalama parametrelerini al
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	// Sıralama ve filtreleme parametrelerini al
	fullName := c.DefaultQuery("fullName", "")
	sortBy := c.DefaultQuery("sortBy", "created_at")
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

	// Cache identifier oluştur - tüm parametreleri içerir
	cacheIdentifier := fmt.Sprintf("applications:list:p%d:l%d:fn%s:s%s:o%s:st%s:e%s:sd%s:ed%s:jid%s",
		page, limit, fullName, sortBy, sortOrder, status, email, startDate, endDate, jobIDStr)

	// Cache kontrolü - önbellekte varsa doğrudan dön
	if h.Cache.TryCache(c, cache.GroupJobs, cacheIdentifier) {
		return
	}

	// Parametreleri SearchParams yapısına dönüştür
	params := types.JobApplicationSearchParams{
		JobID:     jobID,
		Status:    status,
		FullName:  fullName,
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

	// Yanıt hazırla
	response := gin.H{
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
	}

	// Yanıtı önbelleğe al - başvuru listeleri için daha kısa TTL kullanıyoruz
	// çünkü yeni başvurular sık eklenebilir
	h.Cache.SaveCacheTTL(response, cache.GroupJobs, cacheIdentifier, 3*60*time.Second) // 3 dakika

	// Cache header'ı ekle
	c.Header("X-Cache", "MISS")

	// Yanıtı döndür
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetJobApplication(c *gin.Context) {
	// Başvuru ID'sini al
	applicationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "Geçersiz başvuru ID'si")
		return
	}

	// Cache identifier oluştur
	cacheIdentifier := "application:detail:" + applicationID.String()

	// Cache kontrolü - önbellekte varsa doğrudan dön
	if h.Cache.TryCache(c, cache.GroupJobs, cacheIdentifier) {
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

	// Yanıt hazırla
	response := gin.H{
		"success": true,
		"data":    application,
	}

	// Yanıtı önbelleğe al
	h.Cache.SaveCache(response, cache.GroupJobs, cacheIdentifier)

	// Cache header'ı ekle
	c.Header("X-Cache", "MISS")

	// Yanıtı döndür
	c.JSON(http.StatusOK, response)
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

	// Cache identifier oluştur
	cacheIdentifier := "applications:email:" + email

	// Cache kontrolü - önbellekte varsa doğrudan dön
	if h.Cache.TryCache(c, cache.GroupJobs, cacheIdentifier) {
		return
	}

	// Başvuruları getir
	applications, err := h.JobRepository.GetJobApplicationsByEmail(c.Request.Context(), email)
	if err != nil {
		utils.HandleDatabaseError(c, err, "Başvurular getirme")
		return
	}

	// Yanıt hazırla
	response := gin.H{
		"success": true,
		"data":    applications,
	}

	// Yanıtı önbelleğe al - kullanıcı kendi başvurularını görmek için sık sık kontrol edebilir
	h.Cache.SaveCacheTTL(response, cache.GroupJobs, cacheIdentifier, 5*60*time.Second) // 5 dakika

	// Cache header'ı ekle
	c.Header("X-Cache", "MISS")

	// Yanıtı döndür
	c.JSON(http.StatusOK, response)
}
