package JobRepository

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

func (r *Repository) UpdateJob(ctx context.Context, jobID uuid.UUID, input types.JobInput, userID uuid.UUID) (types.Job, error) {
	defer utils.TimeTrack(time.Now(), "Job -> Update Job")
	var job types.Job

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return job, fmt.Errorf("işlem başlatılamadı: %w", err)
	}
	defer tx.Rollback()

	// İş ilanı temel bilgilerini güncelle
	query := `
		UPDATE job_postings
		SET slug = $1, status = $2, deadline = $3, updated_at = NOW()
		WHERE id = $4 AND user_id = $5 AND status != 'deleted'
		RETURNING id, user_id, slug, status, deadline, created_at, updated_at
	`

	err = tx.QueryRowContext(
		ctx,
		query,
		input.Slug,
		input.Status,
		input.Deadline,
		jobID,
		userID,
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
		if err == sql.ErrNoRows {
			return job, fmt.Errorf("güncellenecek iş ilanı bulunamadı veya yetkiniz yok")
		}
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" && pgErr.Constraint == "job_postings_slug_key" {
			return job, fmt.Errorf("bu URL yapısı (%s) zaten kullanımda", input.Slug)
		}
		return job, fmt.Errorf("iş ilanı güncellenemedi: %w", err)
	}

	// İş ilanı detaylarını güncelle
	detailsQuery := `
		UPDATE job_posting_details
		SET title = $1, description = $2, image = $3, location = $4, work_mode = $5, employment_type = $6,
			experience_level = $7, html = $8, json = $9, form_type = $10
		WHERE id = $11
		RETURNING id, title, description, image, location, work_mode, employment_type,
			experience_level, html, json, form_type, applicants
	`

	var details types.JobDetails
	err = tx.QueryRowContext(
		ctx,
		detailsQuery,
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
		jobID,
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
		return job, fmt.Errorf("iş ilanı detayları güncellenemedi: %w", err)
	}

	// Kategorileri güncelle - önce mevcut ilişkileri temizle
	_, err = tx.ExecContext(ctx, "DELETE FROM job_posting_categories WHERE job_id = $1", jobID)
	if err != nil {
		return job, fmt.Errorf("kategoriler temizlenemedi: %w", err)
	}

	// Yeni kategorileri ekle - CreateJob'daki doğru implementasyonu kullanarak
	if len(input.Categories) > 0 {
		categoryQuery := `INSERT INTO job_posting_categories (job_id, category_name) VALUES `

		values := []any{jobID}
		placeholders := []string{}

		for i, category := range input.Categories {
			placeholders = append(placeholders, fmt.Sprintf("($1, $%d)", i+2))
			values = append(values, category)
		}

		// strings.Join kullanarak tüm placeholders'ları virgülle birleştir
		categoryQuery += " " + strings.Join(placeholders, ", ")

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

// UpdateJobStatus sadece iş ilanı durumunu günceller
func (r *Repository) UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status types.JobStatus) error {
	defer utils.TimeTrack(time.Now(), "Job -> Update Job Status")

	// Eğer silme işlemi ise SoftDeleteJob'a yönlendir
	if status == types.JobStatusDeleted {
		return r.SoftDeleteJob(ctx, jobID)
	}

	// Normal durum güncelleme işlemi
	query := `
		UPDATE job_postings
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, status, jobID)
	if err != nil {
		return fmt.Errorf("iş ilanı durumu güncellenemedi: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("etkilenen satır sayısı alınamadı: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("güncellenecek iş ilanı bulunamadı")
	}

	return nil
}
