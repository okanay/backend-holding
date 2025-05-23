package ContentRepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (r *Repository) UpdateContent(ctx context.Context, contentID uuid.UUID, input types.ContentInput, userID uuid.UUID) (types.Content, error) {
	defer utils.TimeTrack(time.Now(), "Repository -> UpdateContent")

	var pc types.Content
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return pc, fmt.Errorf("tx begin error: %w", err)
	}
	defer tx.Rollback()

	var setClauses []string
	var args []any
	paramIndex := 1

	if input.Slug != "" {
		setClauses = append(setClauses, fmt.Sprintf("slug = $%d", paramIndex))
		args = append(args, input.Slug)
		paramIndex++
	}
	if input.Title != "" {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", paramIndex))
		args = append(args, input.Title)
		paramIndex++
	}
	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", paramIndex))
		args = append(args, *input.Description)
		paramIndex++
	}
	if input.ImageURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("image_url = $%d", paramIndex))
		args = append(args, *input.ImageURL)
		paramIndex++
	}

	var detailsJSONStringToUpdate sql.NullString
	if input.DetailsJSON != nil {
		detailsBytes, errJsonMarshal := json.Marshal(input.DetailsJSON)
		if errJsonMarshal != nil {
			return pc, fmt.Errorf("details_json marshal error: %w", errJsonMarshal)
		}
		detailsJSONStringToUpdate.String = string(detailsBytes)
		detailsJSONStringToUpdate.Valid = true
		setClauses = append(setClauses, fmt.Sprintf("details_json = $%d", paramIndex))
		args = append(args, detailsJSONStringToUpdate)
		paramIndex++
	}

	if input.ContentJSON != nil {
		contentBytes, errJsonMarshal := json.Marshal(input.ContentJSON)
		if errJsonMarshal != nil {
			return pc, fmt.Errorf("content_json marshal error: %w", errJsonMarshal)
		}
		setClauses = append(setClauses, fmt.Sprintf("content_json = $%d", paramIndex))
		args = append(args, string(contentBytes))
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

	if len(setClauses) == 0 {
		return r.GetContentByID(ctx, contentID)
	}

	query := fmt.Sprintf(`
		UPDATE contents SET %s
		WHERE id = $%d AND user_id = $%d AND status != '%s'
		RETURNING id, user_id, slug, identifier, language, title, description,
			image_url, details_json, content_json, content_html, status, created_at, updated_at
	`, strings.Join(setClauses, ", "), paramIndex, paramIndex+1, types.ContentStatusDeleted)
	args = append(args, contentID, userID)

	row := tx.QueryRowContext(ctx, query, args...)
	pc, err = scanContent(row)

	if err != nil {
		if err == sql.ErrNoRows {
			return pc, fmt.Errorf("güncellenecek içerik bulunamadı, yetkiniz yok veya silinmiş (ID: %s)", contentID)
		}
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			if pgErr.Constraint == "uq_slug_language" {
				return pc, fmt.Errorf("bu URL yapısı ('%s') ve dil ('%s') zaten başka bir içerik için kullanımda", input.Slug, input.Language)
			}
			return pc, fmt.Errorf("güncelleme sırasında benzersizlik kısıtlaması ihlali (%s): %w", pgErr.Constraint, err)
		}
		return pc, fmt.Errorf("Repository.UpdateContent: scan error: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return pc, fmt.Errorf("tx commit error: %w", err)
	}

	return pc, nil
}

func (r *Repository) UpdateContentStatus(ctx context.Context, contentID uuid.UUID, newStatus types.ContentStatus) error {
	defer utils.TimeTrack(time.Now(), "Repository -> UpdateContentStatus")

	if newStatus == types.ContentStatusDeleted {
		return r.SoftDeleteContent(ctx, contentID)
	}

	query := `UPDATE contents SET status = $1, updated_at = NOW() WHERE id = $2 AND status != $1 AND status != $3`
	result, err := r.db.ExecContext(ctx, query, newStatus, contentID, types.ContentStatusDeleted)
	if err != nil {
		return fmt.Errorf("Repository.UpdateContentStatus: exec error: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Repository.UpdateContentStatus: rows affected error: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("durumu güncellenecek içerik bulunamadı veya durum zaten aynı (ID: %s)", contentID)
	}

	return nil
}
