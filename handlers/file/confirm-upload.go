// handlers/file/confirm-upload.go
package FileHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (h *Handler) ConfirmUpload(c *gin.Context) {
	var input types.ConfirmUploadInput
	err := utils.ValidateRequest(c, &input)
	if err != nil {
		return
	}

	// SignatureID'yi UUID'ye dönüştür
	signatureID, err := uuid.Parse(input.SignatureID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_signature_id",
			"message": "Geçersiz imza ID'si",
		})
		return
	}

	signature, err := h.FileRepository.GetSignatureByID(c.Request.Context(), signatureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "signature_fetch_failed",
			"message": "İmza bilgileri alınamadı: " + err.Error(),
		})
		return
	}

	if signature == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "signature_not_found",
			"message": "İmza kaydı bulunamadı",
		})
		return
	}

	// Dosya kategorisini al
	fileCategory := signature.FileCategory
	if input.FileCategory != "" {
		// Eğer input'da belirtilmişse, onu kullan
		fileCategory = input.FileCategory
	}

	// Dosyayı veritabanına kaydet
	fileInput := types.SaveFileInput{
		URL:          input.URL,
		Filename:     signature.Filename,
		FileType:     signature.FileType,
		FileCategory: fileCategory,
		SizeInBytes:  input.SizeInBytes,
	}

	fileID, err := h.FileRepository.SaveFile(c.Request.Context(), fileInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "file_save_failed",
			"message": "Dosya kaydedilemedi: " + err.Error(),
		})
		return
	}

	// İmza kaydını tamamlandı olarak işaretle
	err = h.FileRepository.CompleteUploadSignature(c.Request.Context(), signatureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "signature_update_failed",
			"message": "İmza kaydı güncellenemedi: " + err.Error(),
		})
		return
	}

	// Başarılı yanıt döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":  fileID.String(),
			"url": input.URL,
		},
	})
}
