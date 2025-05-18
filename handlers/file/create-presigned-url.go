// handlers/file/create-presigned-url.go
package FileHandler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

// CreatePresignedURL dosya yüklemek için presigned URL oluşturur
func (h *Handler) CreatePresignedURL(c *gin.Context) {
	// DEBUG: Print incoming request context and headers
	fmt.Printf("CreatePresignedURL called. Headers: %+v\n", c.Request.Header)

	var input types.CreatePresignedURLInput
	err := utils.ValidateRequest(c, &input)
	if err != nil {
		fmt.Printf("ValidateRequest error: %v\n", err)
		return
	}

	fmt.Printf("Validated input: %+v\n", input)

	// Presigned URL oluştur
	presignedOutput, err := h.R2Repository.GeneratePresignedURL(c.Request.Context(), types.PresignURLInput{
		Filename:     input.Filename,
		ContentType:  input.ContentType,
		FileCategory: input.FileCategory,
		SizeInBytes:  input.SizeInBytes,
	})

	if err != nil {
		fmt.Printf("GeneratePresignedURL error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "presigned_url_failed",
			"message": "Yükleme URL'si oluşturulamadı: " + err.Error(),
		})
		return
	}

	fmt.Printf("PresignedOutput: %+v\n", presignedOutput)

	// Veritabanında signature kaydı oluştur
	signatureInput := types.UploadSignatureInput{
		PresignedURL: presignedOutput.PresignedURL,
		UploadURL:    presignedOutput.UploadURL,
		Filename:     input.Filename,
		FileType:     input.ContentType,
		FileCategory: input.FileCategory,
		ExpiresAt:    presignedOutput.ExpiresAt,
	}

	fmt.Printf("SignatureInput: %+v\n", signatureInput)

	signatureID, err := h.FileRepository.CreateUploadSignature(c.Request.Context(), signatureInput)
	if err != nil {
		fmt.Printf("CreateUploadSignature error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "signature_creation_failed",
			"message": "Yükleme kaydı oluşturulamadı: " + err.Error(),
		})
		return
	}

	fmt.Printf("SignatureID: %v\n", signatureID)

	// Başarılı yanıt döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": types.CreatePresignedURLResponse{
			ID:           signatureID.String(),
			PresignedURL: presignedOutput.PresignedURL,
			UploadURL:    presignedOutput.UploadURL,
			ExpiresAt:    presignedOutput.ExpiresAt,
			Filename:     input.Filename,
		},
	})
}
