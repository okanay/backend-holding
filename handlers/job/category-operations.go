package JobHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (h *Handler) CreateJobCategory(c *gin.Context) {
	// Kullanıcı ID'sini al
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "Bu işlem için giriş yapmanız gerekiyor")
		return
	}

	// İstek verilerini doğrula
	var input types.JobCategoryInput
	if err := utils.ValidateRequest(c, &input); err != nil {
		return
	}

	// Kategoriyi oluştur
	category, err := h.JobRepository.CreateCategory(c.Request.Context(), input, userID.(uuid.UUID))
	if err != nil {
		utils.HandleDatabaseError(c, err, "Kategori oluşturma")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Kategori başarıyla oluşturuldu",
		"data":    category,
	})
}

func (h *Handler) ListJobCategories(c *gin.Context) {
	// Kategorileri getir
	categories, err := h.JobRepository.GetAllCategories(c.Request.Context())
	if err != nil {
		utils.HandleDatabaseError(c, err, "Kategorileri listeleme")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    categories,
	})
}

func (h *Handler) UpdateJobCategory(c *gin.Context) {
	// Kategori adını al
	categoryName := c.Param("name")
	if categoryName == "" {
		utils.BadRequest(c, "Geçersiz kategori adı")
		return
	}

	// İstek verilerini doğrula
	var input types.JobCategoryInput
	if err := utils.ValidateRequest(c, &input); err != nil {
		return
	}

	// Kategoriyi güncelle
	category, err := h.JobRepository.UpdateCategory(c.Request.Context(), categoryName, input)
	if err != nil {
		utils.HandleDatabaseError(c, err, "Kategori güncelleme")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Kategori başarıyla güncellendi",
		"data":    category,
	})
}
