// repositories/r2/presigned.go
package R2Repository

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

// GeneratePresignedURL dosya yüklemek için presigned URL oluşturur
func (r *Repository) GeneratePresignedURL(ctx context.Context, input types.PresignURLInput) (*types.PresignedURLOutput, error) {
	// Dosya adı ve uzantısını ayır
	filename := input.Filename
	fileExt := ""
	dotIndex := strings.LastIndex(filename, ".")

	if dotIndex != -1 {
		fileExt = filename[dotIndex:]  // .jpg, .png vb.
		filename = filename[:dotIndex] // uzantısız dosya adı
	}

	safeFilename := sanitizeFilename(input.Filename)

	// Rastgele hash oluştur (6 karakter)
	hashSuffix := utils.GenerateRandomString(8)

	// Final dosya adını oluştur: orijinal-dosya-adi-ABCDEF.jpg
	finalFilename := fmt.Sprintf("%s-%s%s", safeFilename, hashSuffix, fileExt)

	// File category'ye göre klasör yolu oluştur
	var objectPath string
	if input.FileCategory != "" {
		objectPath = path.Join(r.folderName, input.FileCategory, finalFilename)
	} else {
		objectPath = path.Join(r.folderName, "general", finalFilename)
	}

	// Presigned URL için client oluştur
	presignClient := s3.NewPresignClient(r.client)

	// Presigned URL oluştur
	putObjectRequest, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r.bucketName),
		Key:           aws.String(objectPath),
		ContentType:   aws.String(input.ContentType),
		ContentLength: &input.SizeInBytes,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Minute * 5
	})

	if err != nil {
		return nil, fmt.Errorf("presigned URL oluşturulamadı: %w", err)
	}

	// Public erişim URL'sini oluştur
	publicURL := fmt.Sprintf("%s/%s", r.publicURLBase, objectPath)

	return &types.PresignedURLOutput{
		PresignedURL: putObjectRequest.URL,
		UploadURL:    publicURL,
		ObjectKey:    objectPath,
		ExpiresAt:    time.Now().Add(time.Minute * 5),
	}, nil
}

// Dosya adını güvenli hale getiren yardımcı fonksiyon
func sanitizeFilename(filename string) string {
	// Boşlukları tire ile değiştir
	sanitized := strings.ReplaceAll(filename, " ", "-")

	// Sadece alfanumerik, nokta, tire ve alt çizgi karakterlerine izin ver
	sanitized = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, sanitized)

	return sanitized
}
