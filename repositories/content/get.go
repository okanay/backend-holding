package ContentRepository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

// GetContentByID - ID'ye göre tek bir içerik getirir
func (r *Repository) GetContentByID(ctx context.Context, id uuid.UUID) (types.Content, error) {
	defer utils.TimeTrack(time.Now(), "Repository -> GetContentByID")

	var content types.Content

	query := `
		SELECT
			id, user_id, slug, identifier, language, title, description,
			category, image_url, details_json, content_json, content_html,
			status, created_at, updated_at
		FROM contents
		WHERE id = $1 AND status != $2
		LIMIT 1
	`

	err := r.db.QueryRowContext(ctx, query, id, types.ContentStatusDeleted).Scan(
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
			return content, fmt.Errorf("içerik bulunamadı (ID: %s)", id)
		}
		return content, fmt.Errorf("içerik getirilemedi: %w", err)
	}

	return content, nil
}

// GetContentBySlug - Slug'a göre tüm dillerdeki içerikleri getirir
func (r *Repository) GetContentBySlug(ctx context.Context, slug string) ([]types.Content, error) {
	defer utils.TimeTrack(time.Now(), "Repository -> GetContentBySlug")

	query := `
		WITH target_content AS (
			SELECT identifier
			FROM contents
			WHERE slug = $1 AND status = $2
			LIMIT 1
		)
		SELECT
			c.id,
			c.user_id,
			c.slug,
			c.identifier,
			c.language,
			c.title,
			c.description,
			c.category,
			c.image_url,
			c.details_json,
			c.content_json,
			c.content_html,
			c.status,
			c.created_at,
			c.updated_at
		FROM contents c
		INNER JOIN target_content tc ON c.identifier = tc.identifier
		WHERE c.status = $2
		ORDER BY c.language
	`

	rows, err := r.db.QueryContext(ctx, query, slug, types.ContentStatusPublished)
	if err != nil {
		return nil, fmt.Errorf("sorgu hatası: %w", err)
	}
	defer rows.Close()

	contents, err := scanContents(rows)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

// ListContents - Parametrelere göre içerikleri listeler
func (r *Repository) ListContents(ctx context.Context, params types.ContentSearchParams) ([]types.Content, int, error) {
	defer utils.TimeTrack(time.Now(), "Repository -> ListContents")

	// Base queries
	baseQuery := `
		SELECT
			id, user_id, slug, identifier, language, title, description,
			category, image_url, details_json, content_json, content_html,
			status, created_at, updated_at
		FROM contents
	`
	countQuery := `SELECT COUNT(*) FROM contents`

	// Where clauses
	var whereClauses []string
	var args []any
	paramIndex := 1

	// Status filtresi
	if params.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", paramIndex))
		args = append(args, params.Status)
		paramIndex++
	} else {
		whereClauses = append(whereClauses, fmt.Sprintf("status != $%d", paramIndex))
		args = append(args, types.ContentStatusDeleted)
		paramIndex++
	}

	// Diğer filtreler
	if params.Language != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("language = $%d", paramIndex))
		args = append(args, params.Language)
		paramIndex++
	}

	if params.Identifier != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("identifier = $%d", paramIndex))
		args = append(args, params.Identifier)
		paramIndex++
	}

	if params.Category != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("category = $%d", paramIndex))
		args = append(args, params.Category)
		paramIndex++
	}

	if params.UserID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", paramIndex))
		args = append(args, params.UserID)
		paramIndex++
	}

	// Arama
	if params.Query != "" {
		searchQuery := "%" + strings.ToLower(params.Query) + "%"
		whereClauses = append(whereClauses,
			fmt.Sprintf("(LOWER(title) LIKE $%d OR LOWER(description) LIKE $%d)", paramIndex, paramIndex+1))
		args = append(args, searchQuery, searchQuery)
		paramIndex += 2
	}

	// WHERE clause birleştir
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Sıralama (basit güvenlik kontrolü)
	orderBy := "created_at"
	if params.SortBy != "" {
		allowedSorts := map[string]bool{
			"created_at": true, "updated_at": true, "title": true,
			"status": true, "language": true, "category": true,
		}
		if allowedSorts[params.SortBy] {
			orderBy = params.SortBy
		}
	}

	sortOrder := "DESC"
	if strings.ToUpper(params.SortOrder) == "ASC" {
		sortOrder = "ASC"
	}

	// Sayfalama
	limit := 10
	if params.Limit > 0 && params.Limit <= 100 {
		limit = params.Limit
	}

	page := 1
	if params.Page > 0 {
		page = params.Page
	}
	offset := (page - 1) * limit

	// Toplam sayı
	var total int
	err := r.db.QueryRowContext(ctx, countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("sayım hatası: %w", err)
	}

	if total == 0 {
		return []types.Content{}, 0, nil
	}

	// Ana sorgu
	fullQuery := fmt.Sprintf("%s%s ORDER BY %s %s LIMIT %d OFFSET %d",
		baseQuery, whereClause, orderBy, sortOrder, limit, offset)

	rows, err := r.db.QueryContext(ctx, fullQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listeleme hatası: %w", err)
	}
	defer rows.Close()

	contents, err := scanContents(rows)
	if err != nil {
		return nil, 0, err
	}

	return contents, total, nil
}

// scanContent - Tek bir satırı Content struct'ına dönüştürür
func scanContent(scanner interface{ Scan(dest ...any) error }) (types.Content, error) {
	var content types.Content

	err := scanner.Scan(
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

	return content, err
}

// scanContents - Birden fazla satırı Content slice'ına dönüştürür
func scanContents(rows *sql.Rows) ([]types.Content, error) {
	var contents []types.Content

	for rows.Next() {
		content, err := scanContent(rows)
		if err != nil {
			return nil, fmt.Errorf("satır okuma hatası: %w", err)
		}
		contents = append(contents, content)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("satır iterasyon hatası: %w", err)
	}

	return contents, nil
}
