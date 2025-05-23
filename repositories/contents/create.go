package ContentRepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (r *Repository) CreateContent(ctx context.Context, input types.ContentInput, userID uuid.UUID) (types.Content, error) {
	defer utils.TimeTrack(time.Now(), "Repository -> CreateContent") //

	var pc types.Content
	tx, err := r.db.BeginTx(ctx, nil) //
	if err != nil {
		return pc, fmt.Errorf("tx begin error: %w", err)
	}
	defer tx.Rollback() //

	currentIdentifier := uuid.New()
	if input.Identifier != nil && *input.Identifier != uuid.Nil {
		currentIdentifier = *input.Identifier
	}

	var detailsJSONString sql.NullString
	if input.DetailsJSON != nil {
		detailsBytes, errJsonMarshal := json.Marshal(input.DetailsJSON)
		if errJsonMarshal != nil {
			return pc, fmt.Errorf("details_json marshal error: %w", errJsonMarshal)
		}
		detailsJSONString.String = string(detailsBytes)
		detailsJSONString.Valid = true
	}

	contentBytes, errJsonMarshal := json.Marshal(input.ContentJSON)
	if errJsonMarshal != nil {
		return pc, fmt.Errorf("content_json marshal error: %w", errJsonMarshal)
	}
	contentJSONString := string(contentBytes)

	statusToInsert := types.ContentStatusDraft
	if input.Status != "" {
		statusToInsert = input.Status
	}

	query := `
		INSERT INTO contents (
			user_id, slug, identifier, language, title, description,
			image_url, details_json, content_json, content_html, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING
			id, user_id, slug, identifier, language, title, description,
			image_url, details_json, content_json, content_html, status, created_at, updated_at
	`
	//
	err = tx.QueryRowContext(
		ctx,
		query,
		userID, // Doğrudan UUID, db driver nullability'yi handle etmeli, yoksa sql.NullUUID gerekirdi.
		// types.Content'teki UserID *uuid.UUID olduğu için, userID uuid.Nil ise nil olarak gider.
		input.Slug,
		currentIdentifier,
		input.Language,
		input.Title,
		input.Description, // *string, nil ise NULL olarak gider.
		input.ImageURL,    // *string, nil ise NULL olarak gider.
		detailsJSONString, // sql.NullString
		contentJSONString,
		input.ContentHTML,
		statusToInsert,
	).Scan(
		&pc.ID,
		&pc.UserID, // *uuid.UUID
		&pc.Slug,
		&pc.Identifier,
		&pc.Language,
		&pc.Title,
		&pc.Description, // *string
		&pc.ImageURL,    // *string
		&pc.DetailsJSON, // *string (veritabanından JSON string olarak gelir)
		&pc.ContentJSON, // string (veritabanından JSON string olarak gelir)
		&pc.ContentHTML,
		&pc.Status,
		&pc.CreatedAt,
		&pc.UpdatedAt,
	)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok { //
			if pgErr.Code == "23505" {
				if pgErr.Constraint == "uq_identifier_language" {
					return pc, fmt.Errorf("bu makale için belirtilen dilde ('%s') zaten bir içerik mevcut (identifier: %s)", input.Language, currentIdentifier.String())
				}
				if pgErr.Constraint == "uq_slug_language" {
					return pc, fmt.Errorf("bu URL yapısı ('%s') ve dil ('%s') kombinasyonu zaten kullanımda", input.Slug, input.Language)
				}
				return pc, fmt.Errorf("kayıt zaten mevcut, kısıtlama: %s: %w", pgErr.Constraint, err)
			}
		}
		return pc, fmt.Errorf("Repository.CreateContent: scan error: %w", err)
	}

	if err = tx.Commit(); err != nil { //
		return pc, fmt.Errorf("tx commit error: %w", err)
	}

	return pc, nil
}
