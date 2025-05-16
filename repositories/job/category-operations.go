package JobRepository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (r *Repository) CreateCategory(ctx context.Context, input types.JobCategoryInput, userID uuid.UUID) (*types.JobCategory, error) {
	defer utils.TimeTrack(time.Now(), "Job -> Create Category")

	query := `
		INSERT INTO job_categories (name, display_name, user_id)
		VALUES ($1, $2, $3)
		RETURNING name, display_name, user_id, created_at, updated_at
	`

	var category types.JobCategory
	err := r.db.QueryRowContext(
		ctx,
		query,
		input.Name,
		input.DisplayName,
		userID,
	).Scan(
		&category.Name,
		&category.DisplayName,
		&category.UserID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			if pgErr.Constraint == "job_categories_pkey" || pgErr.Constraint == "job_categories_name_key" {
				return nil, fmt.Errorf("bu kategori adı (%s) zaten kullanımda", input.Name)
			}
		}
		return nil, fmt.Errorf("kategori oluşturulamadı: %w", err)
	}

	return &category, nil
}

func (r *Repository) GetAllCategories(ctx context.Context) ([]types.JobCategory, error) {
	defer utils.TimeTrack(time.Now(), "Job -> Get All Categories")

	query := `
		SELECT name, display_name, user_id, created_at, updated_at
		FROM job_categories
		ORDER BY display_name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("kategoriler getirilemedi: %w", err)
	}
	defer rows.Close()

	var categories []types.JobCategory
	for rows.Next() {
		var category types.JobCategory
		if err := rows.Scan(
			&category.Name,
			&category.DisplayName,
			&category.UserID,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			continue
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("kategori listesi işlenirken hata: %w", err)
	}

	return categories, nil
}

func (r *Repository) UpdateCategory(ctx context.Context, name string, input types.JobCategoryInput) (*types.JobCategory, error) {
	defer utils.TimeTrack(time.Now(), "Job -> Update Category")

	query := `
		UPDATE job_categories
		SET display_name = $1, updated_at = NOW()
		WHERE name = $2
		RETURNING name, display_name, user_id, created_at, updated_at
	`

	var category types.JobCategory
	err := r.db.QueryRowContext(
		ctx,
		query,
		input.DisplayName,
		name,
	).Scan(
		&category.Name,
		&category.DisplayName,
		&category.UserID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("güncellenecek kategori bulunamadı: %s", name)
		}
		return nil, fmt.Errorf("kategori güncellenemedi: %w", err)
	}

	return &category, nil
}
