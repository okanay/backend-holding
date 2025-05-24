package ContentRepository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

// UpdateContent - İçeriği günceller (PATCH mantığı - sadece gönderilen alanları günceller)
func (r *Repository) UpdateContent(ctx context.Context, contentID uuid.UUID, input types.ContentInput, userID uuid.UUID) (types.Content, error) {
	defer utils.TimeTrack(time.Now(), "Repository -> UpdateContent")

	var content types.Content

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return content, fmt.Errorf("transaction başlatılamadı: %w", err)
	}
	defer tx.Rollback()

	// Dinamik UPDATE sorgusu oluştur
	var setClauses []string
	var args []any
	paramIndex := 1

	// Her alan için kontrol ve ekleme
	if input.Category != "" {
		setClauses = append(setClauses, fmt.Sprintf("category = $%d", paramIndex))
		args = append(args, input.Category)
		paramIndex++
	}

	if input.Slug != "" {
		setClauses = append(setClauses, fmt.Sprintf("slug = $%d", paramIndex))
		args = append(args, input.Slug)
		paramIndex++
	}

	if input.Identifier != "" {
		setClauses = append(setClauses, fmt.Sprintf("identifier = $%d", paramIndex))
		args = append(args, input.Identifier)
		paramIndex++
	}

	if input.Language != "" {
		setClauses = append(setClauses, fmt.Sprintf("language = $%d", paramIndex))
		args = append(args, input.Language)
		paramIndex++
	}

	if input.Title != "" {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", paramIndex))
		args = append(args, input.Title)
		paramIndex++
	}

	if input.Description != "" {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", paramIndex))
		args = append(args, input.Description)
		paramIndex++
	}

	if input.ImageURL != "" {
		setClauses = append(setClauses, fmt.Sprintf("image_url = $%d", paramIndex))
		args = append(args, input.ImageURL)
		paramIndex++
	}

	if input.DetailsJSON != "" {
		setClauses = append(setClauses, fmt.Sprintf("details_json = $%d", paramIndex))
		args = append(args, sql.NullString{String: input.DetailsJSON, Valid: true})
		paramIndex++
	}

	if input.ContentJSON != "" {
		setClauses = append(setClauses, fmt.Sprintf("content_json = $%d", paramIndex))
		args = append(args, input.ContentJSON)
		paramIndex++
	}

	if input.ContentHTML != "" {
		setClauses = append(setClauses, fmt.Sprintf("content_html = $%d", paramIndex))
		args = append(args, input.ContentHTML)
		paramIndex++
	}

	if input.Status != "" {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", paramIndex))
		args = append(args, input.Status)
		paramIndex++
	}

	// Güncellenecek alan yoksa mevcut veriyi dön
	if len(setClauses) == 0 {
		return r.GetContentByID(ctx, contentID)
	}

	// updated_at ekle
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", paramIndex))
	args = append(args, time.Now())
	paramIndex++

	// WHERE parametreleri
	args = append(args, contentID, userID, types.ContentStatusDeleted)

	// Sorguyu oluştur ve çalıştır
	query := fmt.Sprintf(`
		UPDATE contents
		SET %s
		WHERE id = $%d AND user_id = $%d AND status != $%d
		RETURNING
			id, user_id, slug, identifier, language, title, description,
			category, image_url, details_json, content_json, content_html,
			status, created_at, updated_at
	`, strings.Join(setClauses, ", "), paramIndex, paramIndex+1, paramIndex+2)

	err = tx.QueryRowContext(ctx, query, args...).Scan(
		&content.ID,
		&content.UserID,
		&content.Slug,
		&content.Identifier,
		&content.Language,
		&content.Title,
		&content.Description,
		&content.Category,
		&content.ImageURL,
		&content.DetailsJSON,
		&content.ContentJSON,
		&content.ContentHTML,
		&content.Status,
		&content.CreatedAt,
		&content.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return content, fmt.Errorf("içerik bulunamadı, yetkiniz yok veya silinmiş (ID: %s)", contentID)
		}

		// PostgreSQL benzersizlik hatası
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			if pgErr.Constraint == "uq_slug_language" {
				return content, fmt.Errorf("'%s' URL'i '%s' dilinde zaten kullanımda", input.Slug, input.Language)
			}
			if pgErr.Constraint == "uq_identifier_language" {
				return content, fmt.Errorf("bu içerik için '%s' dilinde zaten versiyon var", input.Language)
			}
		}

		return content, fmt.Errorf("güncelleme hatası: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return content, fmt.Errorf("transaction commit hatası: %w", err)
	}

	return content, nil
}

// UpdateContentStatus - Sadece içerik durumunu günceller
func (r *Repository) UpdateContentStatus(ctx context.Context, contentID uuid.UUID, newStatus types.ContentStatus) error {
	defer utils.TimeTrack(time.Now(), "Repository -> UpdateContentStatus")

	// Deleted status için soft delete kullan
	if newStatus == types.ContentStatusDeleted {
		return r.SoftDeleteContent(ctx, contentID)
	}

	query := `
		UPDATE contents
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND status != $1 AND status != $3
	`

	result, err := r.db.ExecContext(ctx, query, newStatus, contentID, types.ContentStatusDeleted)
	if err != nil {
		return fmt.Errorf("status güncelleme hatası: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("etkilenen satır sayısı alınamadı: %w", err)
	}

	if rowsAffected == 0 {
		// İçerik var mı kontrol et
		var exists bool
		err = r.db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM contents WHERE id = $1)",
			contentID).Scan(&exists)

		if err != nil {
			return fmt.Errorf("varlık kontrolü hatası: %w", err)
		}

		if !exists {
			return fmt.Errorf("içerik bulunamadı (ID: %s)", contentID)
		}

		return fmt.Errorf("içerik durumu zaten '%s' veya silinmiş", newStatus)
	}

	return nil
}
