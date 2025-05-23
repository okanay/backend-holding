package ContentHandler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (h *Handler) GetContentByID(c *gin.Context) {
	contentIDStr := c.Param("id")
	contentID, err := uuid.Parse(contentIDStr)
	if err != nil {
		utils.BadRequest(c, "Geçersiz içerik ID'si")
		return
	}

	cacheIdentifier := fmt.Sprintf("content:detail:%s", contentIDStr)
	if h.Cache.TryCache(c, Group, cacheIdentifier) {
		return
	}

	pc, err := h.Repository.GetContentByID(c.Request.Context(), contentID)
	if err != nil {
		if err.Error() == fmt.Sprintf("içerik bulunamadı (ID: %s)", contentIDStr) {
			utils.NotFound(c, "Basın içeriği")
			return
		}
		utils.HandleDatabaseError(c, err, "Basın içeriği getirme")
		return
	}

	response := gin.H{
		"success": true,
		"data":    mapContentToView(pc),
	}

	h.Cache.SaveCache(response, Group, cacheIdentifier)
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetNewsBySlug(c *gin.Context) {
	slug := strings.ToLower(c.Param("slug"))
	requestedLang := strings.ToLower(c.Param("lang"))

	// Input validation
	if slug == "" {
		utils.BadRequest(c, "Geçersiz URL yapısı (slug eksik)")
		return
	}
	if requestedLang == "" {
		utils.BadRequest(c, "Geçersiz URL yapısı (dil kodu eksik)")
		return
	}

	// Cache kontrolü - requestedLang'i de key'e ekledik
	cacheIdentifier := fmt.Sprintf("content:slug_response:%s:%s", slug, requestedLang)
	if h.Cache.TryCache(c, Group, cacheIdentifier) {
		return
	}

	// Veritabanından tüm dillerdeki içerikleri getir
	allLanguageContents, err := h.Repository.GetNewsBySlug(c.Request.Context(), slug)
	if err != nil {
		utils.HandleDatabaseError(c, err, "Basın içerikleri (slug ile) getirme")
		return
	}

	if len(allLanguageContents) == 0 {
		utils.NotFound(c, "Belirtilen slug ile ilişkili basın içeriği bulunamadı")
		return
	}

	// Published content'leri filtrele ve organize et
	publishedContents := filterPublishedContents(allLanguageContents)
	if len(publishedContents) == 0 {
		utils.NotFound(c, "Yayınlanmış içerik bulunamadı")
		return
	}

	// İstenen dildeki content'i bul
	requestedLangContent := findContentByLanguage(publishedContents, requestedLang)
	isCanonical := requestedLangContent.Slug == slug
	canonicalData := mapContentToView(*requestedLangContent)
	alternateDatas := buildAlternateLanguages(publishedContents, requestedLang)

	finalResponse := gin.H{
		"success": true,
		"data": gin.H{
			"isCanonical":    isCanonical,
			"canonicalData":  canonicalData,
			"haveAlternate":  len(alternateDatas) > 0,
			"alternateDatas": alternateDatas,
		},
	}

	// Cache'e kaydet ve döndür
	h.Cache.SaveCache(finalResponse, Group, cacheIdentifier)
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, finalResponse)
}

func (h *Handler) ListContents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	sortBy := c.DefaultQuery("sortBy", "createdAt")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	status := types.ContentStatus(c.Query("status"))
	language := c.Query("language")
	identifier := c.Query("identifier")
	query := c.Query("q")
	userID := c.Query("userId")

	cacheIdentifier := fmt.Sprintf("content:list:p%d:l%d:s%s:o%s:st%s:lang%s:id%s:q%s:uid%s",
		page, limit, sortBy, sortOrder, status, language, identifier, query, userID)

	if h.Cache.TryCache(c, Group, cacheIdentifier) {
		return
	}

	params := types.ContentSearchParams{
		Status:     status,
		Language:   language,
		Identifier: identifier,
		Query:      query,
		UserID:     userID,
		Page:       page,
		Limit:      limit,
		SortBy:     sortBy,
		SortOrder:  sortOrder,
	}

	contents, total, err := h.Repository.ListContents(c.Request.Context(), params)
	if err != nil {
		utils.HandleDatabaseError(c, err, "Basın içerikleri listeleme")
		return
	}

	response := gin.H{
		"success": true,
		"data": gin.H{
			"contents": mapContentsToViews(contents),
			"pagination": gin.H{
				"currentPage": page,
				"pageSize":    limit,
				"totalItems":  total,
				"totalPages":  (total + limit - 1) / limit,
			},
		},
	}

	h.Cache.SaveCacheTTL(response, Group, cacheIdentifier, 5*time.Minute)
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, response)
}

func (h *Handler) ListPublishedContents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	sortBy := c.DefaultQuery("sortBy", "createdAt")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	language := c.Query("language")
	identifier := c.Query("identifier")
	query := c.Query("q")

	cacheIdentifier := fmt.Sprintf("content:published:p%d:l%d:s%s:o%s:lang%s:id%s:q%s",
		page, limit, sortBy, sortOrder, language, identifier, query)

	if h.Cache.TryCache(c, Group, cacheIdentifier) {
		return
	}

	params := types.ContentSearchParams{
		Status:     types.ContentStatusPublished,
		Language:   language,
		Identifier: identifier,
		Query:      query,
		Page:       page,
		Limit:      limit,
		SortBy:     sortBy,
		SortOrder:  sortOrder,
	}

	contents, total, err := h.Repository.ListContents(c.Request.Context(), params)
	if err != nil {
		utils.HandleDatabaseError(c, err, "Yayınlanmış basın içerikleri listeleme")
		return
	}

	response := gin.H{
		"success": true,
		"data": gin.H{
			"contents": mapContentsToViews(contents),
			"pagination": gin.H{
				"currentPage": page,
				"pageSize":    limit,
				"totalItems":  total,
				"totalPages":  (total + limit - 1) / limit,
			},
		},
	}
	h.Cache.SaveCacheTTL(response, Group, cacheIdentifier, 15*time.Minute)
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, response)
}
