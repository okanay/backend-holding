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

func (r *Repository) GetContentByID(ctx context.Context, id uuid.UUID) (types.Content, error) {
	defer utils.TimeTrack(time.Now(), "Repository -> GetContentByID")
	var pc types.Content
	query := `
		SELECT id, user_id, slug, identifier, language, title, description,
						image_url, details_json, content_json, content_html, status, created_at, updated_at
		FROM press_contents
		WHERE id = $1 AND status != 'deleted' LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, id)
	pc, err := scanContent(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return pc, fmt.Errorf("içerik bulunamadı (ID: %s)", id)
		}
		return pc, fmt.Errorf("Repository.GetContentByID: scan error: %w", err)
	}
	return pc, nil
}

func (r *Repository) GetNewsBySlug(ctx context.Context, slug string) ([]types.Content, error) {
	defer utils.TimeTrack(time.Now(), "Repository -> GetNewsBySlug")

	query := `
		SELECT
			pc.id,
			pc.user_id,
			pc.slug,
			pc.identifier,
			pc.language,
			pc.title,
			pc.description,
			pc.image_url,
			pc.details_json,
			pc.content_json,
			pc.content_html,
			pc.status,
			pc.created_at,
			pc.updated_at
		FROM
			press_contents pc
		WHERE
			pc.identifier = (
				SELECT sub_pc.identifier
				FROM press_contents sub_pc
				WHERE sub_pc.slug = $1
				AND sub_pc.status = 'published'
				LIMIT 1
			)
		AND pc.status = 'published'
		ORDER BY pc.language;
	`

	rows, err := r.db.QueryContext(ctx, query, slug)
	if err != nil {
		return nil, fmt.Errorf("Repository.GetNewsBySlug: query error: %w", err)
	}
	defer rows.Close()

	contents, err := scanContents(rows)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

func (r *Repository) ListContents(ctx context.Context, params types.ContentSearchParams) ([]types.Content, int, error) {
	defer utils.TimeTrack(time.Now(), "Repository -> ListContents") //

	baseQuery := `
		SELECT id, user_id, slug, identifier, language, title, description,
			   image_url, details_json, content_json, content_html, status, created_at, updated_at
		FROM press_contents
	` //
	countQuery := `SELECT COUNT(*) FROM press_contents` //

	var whereClauses []string
	var args []any
	paramIndex := 1

	if params.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", paramIndex))
		args = append(args, params.Status)
		paramIndex++
	} else {
		whereClauses = append(whereClauses, fmt.Sprintf("status != $%d", paramIndex))
		args = append(args, types.ContentStatusDeleted)
		paramIndex++
	}

	if params.Language != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("language = $%d", paramIndex))
		args = append(args, params.Language)
		paramIndex++
	}

	if params.Identifier != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("identifier = $%d", paramIndex))
		args = append(args, params.Identifier) // types.ContentSearchParams.Identifier string
		paramIndex++
	}

	if params.UserID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", paramIndex))
		args = append(args, params.UserID) // types.ContentSearchParams.UserID string
		paramIndex++
	}

	if params.Query != "" {
		searchQuery := "%" + strings.ToLower(params.Query) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(title) LIKE $%d OR LOWER(description) LIKE $%d)", paramIndex, paramIndex+1))
		args = append(args, searchQuery, searchQuery)
		paramIndex += 2
	}

	finalWhereClause := ""
	if len(whereClauses) > 0 {
		finalWhereClause = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	orderBy := "created_at"
	allowedSortBy := map[string]string{
		"createdAt": "created_at", "updatedAt": "updated_at", "title": "title",
		"status": "status", "language": "language",
	}
	if col, ok := allowedSortBy[params.SortBy]; ok { //
		orderBy = col
	}

	sortOrder := "DESC"
	if strings.ToUpper(params.SortOrder) == "DESC" || strings.ToUpper(params.SortOrder) == "ASC" { //
		sortOrder = strings.ToUpper(params.SortOrder)
	}
	orderClause := fmt.Sprintf(" ORDER BY %s %s", orderBy, sortOrder) //

	limit := 10
	if params.Limit > 0 {
		limit = params.Limit
	}
	page := 1
	if params.Page > 0 {
		page = params.Page
	}
	offset := (page - 1) * limit
	paginationClause := fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset) //

	var total int
	err := r.db.QueryRowContext(ctx, countQuery+finalWhereClause, args...).Scan(&total) //
	if err != nil {
		return nil, 0, fmt.Errorf("Repository.ListContents: count query error: %w", err)
	}

	if total == 0 {
		return []types.Content{}, 0, nil
	}

	rows, err := r.db.QueryContext(ctx, baseQuery+finalWhereClause+orderClause+paginationClause, args...) //
	if err != nil {
		return nil, 0, fmt.Errorf("Repository.ListContents: select query error: %w", err)
	}
	defer rows.Close()

	contents, err := scanContents(rows) //
	if err != nil {
		return nil, 0, err // Hata zaten sarılı geldi.
	}

	return contents, total, nil
}

func scanContent(scanner interface{ Scan(dest ...any) error }) (types.Content, error) {
	var pc types.Content
	err := scanner.Scan(
		&pc.ID,
		&pc.UserID,
		&pc.Slug,
		&pc.Identifier,
		&pc.Language,
		&pc.Title,
		&pc.Description,
		&pc.ImageURL,
		&pc.DetailsJSON,
		&pc.ContentJSON,
		&pc.ContentHTML,
		&pc.Status,
		&pc.CreatedAt,
		&pc.UpdatedAt,
	)
	return pc, err
}

func scanContents(rows *sql.Rows) ([]types.Content, error) {
	var contents []types.Content
	for rows.Next() {
		pc, err := scanContent(rows)
		if err != nil {
			return nil, fmt.Errorf("Repository.scanContents: row scan error: %w", err)
		}
		contents = append(contents, pc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Repository.scanContents: rows iteration error: %w", err)
	}
	return contents, nil
}
