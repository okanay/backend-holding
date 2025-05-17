package JobRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (r *Repository) CreateJob(ctx context.Context, input types.JobInput, userID uuid.UUID) (types.Job, error) {
	defer utils.TimeTrack(time.Now(), "Job -> Create Job")
	var job types.Job

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return job, fmt.Errorf("işlem başlatılamadı: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO job_postings (user_id, slug, status, deadline)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, slug, status, deadline, created_at, updated_at
	`

	err = tx.QueryRowContext(
		ctx,
		query,
		userID,
		input.Slug,
		input.Status,
		input.Deadline,
	).Scan(
		&job.ID,
		&job.UserID,
		&job.Slug,
		&job.Status,
		&job.Deadline,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" && pgErr.Constraint == "job_postings_slug_key" {
			return job, fmt.Errorf("bu URL yapısı (%s) zaten kullanımda", input.Slug)
		}
		return job, fmt.Errorf("iş ilanı oluşturulamadı: %w", err)
	}

	detailsQuery := `
		INSERT INTO job_posting_details (
			id, title, description, image, location, work_mode, employment_type,
			experience_level, html, json, form_type, applicants
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 0)
		RETURNING id, title, description, image, location, work_mode, employment_type,
			experience_level, html, json, form_type, applicants
	`

	var details types.JobDetails
	err = tx.QueryRowContext(
		ctx,
		detailsQuery,
		job.ID,
		input.Title,
		input.Description,
		input.Image,
		input.Location,
		input.WorkMode,
		input.EmploymentType,
		input.ExperienceLevel,
		input.HTML,
		input.JSON,
		input.FormType,
	).Scan(
		&details.ID,
		&details.Title,
		&details.Description,
		&details.Image,
		&details.Location,
		&details.WorkMode,
		&details.EmploymentType,
		&details.ExperienceLevel,
		&details.HTML,
		&details.JSON,
		&details.FormType,
		&details.Applicants,
	)

	if err != nil {
		return job, fmt.Errorf("iş ilanı detayları oluşturulamadı: %w", err)
	}

	if len(input.Categories) > 0 {
		categoryQuery := `INSERT INTO job_posting_categories (job_id, category_name) VALUES `

		values := []any{job.ID}
		placeholders := []string{}

		for i, category := range input.Categories {
			placeholders = append(placeholders, fmt.Sprintf("($1, $%d)", i+2))
			values = append(values, category)
		}

		categoryQuery += " " + fmt.Sprint(placeholders)

		_, err = tx.ExecContext(ctx, categoryQuery, values...)
		if err != nil {
			if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23503" {
				return job, fmt.Errorf("belirtilen kategorilerden bazıları bulunamadı: %w", err)
			}
			return job, fmt.Errorf("kategoriler eklenemedi: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return job, fmt.Errorf("işlem tamamlanamadı: %w", err)
	}

	job.Details = &details
	job.Categories = input.Categories

	return job, nil
}
