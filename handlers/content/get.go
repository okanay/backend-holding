package ContentHandler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/services/cache"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

// GetContentByID - ID ile içerik getirir
func (h *Handler) GetContentByID(c *gin.Context) {
	// ID'yi parse et
	contentIDStr := c.Param("id")
	contentID, err := uuid.Parse(contentIDStr)
	if err != nil {
		utils.BadRequest(c, "Geçersiz içerik ID'si")
		return
	}

	// Cache kontrolü
	cacheKey := fmt.Sprintf("content:id:%s", contentIDStr)
	if h.Cache.TryCache(c, cache.GroupContent, cacheKey) {
		return
	}

	// Veritabanından getir
	content, err := h.Repository.GetContentByID(c.Request.Context(), contentID)
	if err != nil {
		if strings.Contains(err.Error(), "bulunamadı") {
			utils.NotFound(c, "İçerik bulunamadı")
			return
		}
		utils.HandleDatabaseError(c, err, "İçerik getirme")
		return
	}

	// Response oluştur
	response := gin.H{
		"success": true,
		"data":    mapContentToView(content),
	}

	// Cache'e kaydet ve dön
	h.Cache.SaveCache(response, cache.GroupContent, cacheKey)
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, response)
}

// GetContentBySlug - Slug ve dil ile içerik getirir
func (h *Handler) GetContentBySlug(c *gin.Context) {
	slug := strings.ToLower(c.Param("slug"))
	lang := strings.ToLower(c.Param("lang"))

	// Validasyon
	if slug == "" || lang == "" {
		utils.BadRequest(c, "Slug ve dil parametreleri zorunludur")
		return
	}

	// Cache kontrolü
	cacheKey := fmt.Sprintf("content:slug:%s:%s", slug, lang)
	if h.Cache.TryCache(c, cache.GroupContent, cacheKey) {
		return
	}

	// Tüm dillerdeki versiyonları getir
	contents, err := h.Repository.GetContentBySlug(c.Request.Context(), slug)
	if err != nil {
		utils.HandleDatabaseError(c, err, "İçerik getirme")
		return
	}

	if len(contents) == 0 {
		utils.NotFound(c, "İçerik bulunamadı")
		return
	}

	// İstenen dili bul
	var requestedContent *types.Content
	var alternates []gin.H

	for _, content := range contents {
		if strings.ToLower(content.Language) == lang {
			requestedContent = &content
		} else {
			alternates = append(alternates, gin.H{
				"language": content.Language,
				"slug":     content.Slug,
				"title":    content.Title,
			})
		}
	}

	if requestedContent == nil {
		utils.NotFound(c, fmt.Sprintf("İçerik '%s' dilinde bulunamadı", lang))
		return
	}

	// Response oluştur
	response := gin.H{
		"success": true,
		"data": gin.H{
			"content":    mapContentToView(*requestedContent),
			"alternates": alternates,
		},
	}

	// Cache ve dön
	h.Cache.SaveCache(response, cache.GroupContent, cacheKey)
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, response)
}

// ListContents - İçerikleri listeler (admin için)
func (h *Handler) ListContents(c *gin.Context) {
	// Query parametrelerini al
	params := types.ContentSearchParams{
		Page:       parseInt(c.DefaultQuery("page", "1")),
		Limit:      parseInt(c.DefaultQuery("limit", "10")),
		SortBy:     c.DefaultQuery("sortBy", "created_at"),
		SortOrder:  c.DefaultQuery("sortOrder", "desc"),
		Status:     types.ContentStatus(c.Query("status")),
		Language:   c.Query("language"),
		Category:   c.Query("category"),
		Identifier: c.Query("identifier"),
		Query:      c.Query("q"),
		UserID:     c.Query("userId"),
	}

	// Cache key
	cacheKey := fmt.Sprintf("content:list:%+v", params)
	if h.Cache.TryCache(c, cache.GroupContent, cacheKey) {
		return
	}

	// Veritabanından getir
	contents, total, err := h.Repository.ListContents(c.Request.Context(), params)
	if err != nil {
		utils.HandleDatabaseError(c, err, "İçerik listeleme")
		return
	}

	// Response
	response := gin.H{
		"success": true,
		"data": gin.H{
			"contents": mapContentsToViews(contents),
			"pagination": gin.H{
				"page":       params.Page,
				"limit":      params.Limit,
				"total":      total,
				"totalPages": (total + params.Limit - 1) / params.Limit,
			},
		},
	}

	// Cache ve dön
	h.Cache.SaveCacheTTL(response, cache.GroupContent, cacheKey, 5*time.Minute)
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, response)
}

// ListPublishedContents - Yayınlanmış içerikleri listeler (public için)
func (h *Handler) ListPublishedContents(c *gin.Context) {
	// Query parametreleri
	params := types.ContentSearchParams{
		Page:      parseInt(c.DefaultQuery("page", "1")),
		Limit:     parseInt(c.DefaultQuery("limit", "10")),
		SortBy:    c.DefaultQuery("sortBy", "created_at"),
		SortOrder: c.DefaultQuery("sortOrder", "desc"),
		Status:    types.ContentStatusPublished, // Sadece yayınlanmış
		Language:  c.Query("language"),
		Category:  c.Query("category"),
		Query:     c.Query("q"),
	}

	// Cache key
	cacheKey := fmt.Sprintf("content:published:%+v", params)
	if h.Cache.TryCache(c, cache.GroupContent, cacheKey) {
		return
	}

	// Veritabanından getir
	contents, total, err := h.Repository.ListContents(c.Request.Context(), params)
	if err != nil {
		utils.HandleDatabaseError(c, err, "İçerik listeleme")
		return
	}

	// Response
	response := gin.H{
		"success": true,
		"data": gin.H{
			"contents": mapContentsToViews(contents),
			"pagination": gin.H{
				"page":       params.Page,
				"limit":      params.Limit,
				"total":      total,
				"totalPages": (total + params.Limit - 1) / params.Limit,
			},
		},
	}

	// Cache ve dön
	h.Cache.SaveCacheTTL(response, cache.GroupContent, cacheKey, 15*time.Minute)
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, response)
}

// Helper function
func parseInt(s string) int {
	val, _ := strconv.Atoi(s)
	if val <= 0 {
		return 1
	}
	return val
}
