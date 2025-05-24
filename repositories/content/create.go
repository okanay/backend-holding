package ContentRepository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

// CreateContent - Yeni içerik oluşturur
func (r *Repository) CreateContent(ctx context.Context, input types.ContentInput, userID uuid.UUID) (types.Content, error) {
	defer utils.TimeTrack(time.Now(), "Repository -> CreateContent")

	var content types.Content

	// Transaction başlat
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return content, fmt.Errorf("transaction başlatılamadı: %w", err)
	}
	defer tx.Rollback()

	// ContentJSON'ı hazırla (zorunlu alan)
	contentJSONBytes, err := json.Marshal(input.ContentJSON)
	if err != nil {
		return content, fmt.Errorf("content_json hazırlanamadı: %w", err)
	}

	// details_json'ı hazırla (JSONB formatında)
	detailsJSONBytes, err := json.Marshal(input.DetailsJSON)
	if err != nil {
		return content, fmt.Errorf("details_json hazırlanamadı: %w", err)
	}

	// Status belirle - default: draft
	status := types.ContentStatusDraft
	if input.Status != "" {
		status = input.Status
	}

	// INSERT sorgusu
	query := `
		INSERT INTO contents (
			user_id, slug, identifier, language, title, description,
			category, image_url, details_json, content_json, content_html, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		) RETURNING
			id, user_id, slug, identifier, language, title, description,
			category, image_url, details_json, content_json, content_html,
			status, created_at, updated_at
	`

	// Sorguyu çalıştır
	err = tx.QueryRowContext(ctx, query,
		userID,
		input.Slug,
		input.Identifier,
		input.Language,
		input.Title,
		input.Description,
		input.Category,
		input.ImageURL,
		string(detailsJSONBytes),
		string(contentJSONBytes),
		input.ContentHTML,
		status,
	).Scan(
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
		// PostgreSQL hata kontrolü
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			// Benzersizlik hatası
			switch pgErr.Constraint {
			case "uq_identifier_language":
				return content, fmt.Errorf("bu içerik için '%s' dilinde zaten bir versiyon mevcut", input.Language)
			case "uq_slug_language":
				return content, fmt.Errorf("'%s' URL'i '%s' dilinde zaten kullanımda", input.Slug, input.Language)
			default:
				return content, fmt.Errorf("benzersizlik hatası: %s", pgErr.Detail)
			}
		}
		return content, fmt.Errorf("içerik oluşturulamadı: %w", err)
	}

	// Transaction'ı commit et
	if err = tx.Commit(); err != nil {
		return content, fmt.Errorf("transaction commit hatası: %w", err)
	}

	return content, nil
}
