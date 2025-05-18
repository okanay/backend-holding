package JobRepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

// getJobQuery - Tüm iş ilanı sorgularında kullanılacak ortak SQL sorgusu
const jobBaseQuery = `
				SELECT
								p.id,
								p.slug,
								p.status,
								p.deadline,
								p.created_at,
								p.updated_at,
								d.title,
								d.description,
								d.image,
								d.location,
								d.work_mode,
								d.employment_type,
								d.experience_level,
								d.html,
								d.json,
								d.form_type,
								d.applicants,
								-- Kategorileri dizi olarak al
								(
												SELECT COALESCE(json_agg(
																json_build_object(
																				'name', c.category_name,
																				'displayName', cat.display_name,
																				'createdAt', cat.created_at
																) ORDER BY cat.display_name
												), '[]'::json)
												FROM job_posting_categories c
												LEFT JOIN job_categories cat ON c.category_name = cat.name
												WHERE c.job_id = p.id
								) AS categories
				FROM job_postings p
				LEFT JOIN job_posting_details d ON p.id = d.id
`

// scanJob - Ortak scan işlemi için yardımcı fonksiyon
func scanJob(row *sql.Row) (types.JobView, error) {
	var job types.JobView
	var details types.JobDetailsView
	var categoriesJSON []byte

	err := row.Scan(
		&job.ID,
		&job.Slug,
		&job.Status,
		&job.Deadline,
		&job.CreatedAt,
		&job.UpdatedAt,
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
		&categoriesJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return job, nil // İlan bulunamadı
		}
		return job, fmt.Errorf("iş ilanı getirilirken hata: %w", err)
	}

	// Kategorileri ayrıştır
	var categories []types.JobCategoryView
	if err := json.Unmarshal(categoriesJSON, &categories); err != nil {
		return job, fmt.Errorf("kategoriler ayrıştırılamadı: %w", err)
	}

	// JobView nesnesini tamamla
	job.Details = details
	job.Categories = categories

	return job, nil
}

// scanJobs - Ortak rows scan işlemi için yardımcı fonksiyon
func scanJobs(rows *sql.Rows) ([]types.JobView, error) {
	var jobs []types.JobView

	for rows.Next() {
		var job types.JobView
		var details types.JobDetailsView
		var categoriesJSON []byte

		err := rows.Scan(
			&job.ID,
			&job.Slug,
			&job.Status,
			&job.Deadline,
			&job.CreatedAt,
			&job.UpdatedAt,
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
			&categoriesJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("iş ilanı taranırken hata: %w", err)
		}

		// Kategorileri ayrıştır
		var categories []types.JobCategoryView
		if err := json.Unmarshal(categoriesJSON, &categories); err != nil {
			return nil, fmt.Errorf("kategoriler ayrıştırılamadı: %w", err)
		}

		// JobView nesnesini tamamla
		job.Details = details
		job.Categories = categories

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iş ilanları okunurken hata: %w", err)
	}

	return jobs, nil
}

// GetAllJobs - Tüm iş ilanlarını view olarak getirir
func (r *Repository) GetAllJobs(ctx context.Context) ([]types.JobView, error) {
	defer utils.TimeTrack(time.Now(), "Job -> Get All Jobs")

	query := jobBaseQuery + `
								WHERE p.status = 'published'
								ORDER BY p.created_at DESC
				`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("iş ilanları getirilemedi: %w", err)
	}
	defer rows.Close()

	return scanJobs(rows)
}

// ListJobs iş ilanlarını arama parametrelerine göre listeler
func (r *Repository) ListJobs(ctx context.Context, params types.JobSearchParams) ([]types.JobView, int, error) {
	defer utils.TimeTrack(time.Now(), "Job -> List Jobs")

	baseQuery := jobBaseQuery

	countQuery := `
		SELECT COUNT(*)
		FROM job_postings p
		LEFT JOIN job_posting_details d ON p.id = d.id
	`

	whereClause := " WHERE p.status != 'deleted'"
	args := []any{}
	paramIndex := 1

	// Durum filtreleme
	if params.Status != "" {
		whereClause += fmt.Sprintf(" AND p.status = $%d", paramIndex)
		args = append(args, params.Status)
		paramIndex++
	}

	// Kategori filtreleme
	if params.Category != "" {
		whereClause += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM job_posting_categories jpc WHERE jpc.job_id = p.id AND jpc.category_name = $%d)", paramIndex)
		args = append(args, params.Category)
		paramIndex++
	}

	// Lokasyon filtreleme
	if params.Location != "" {
		whereClause += fmt.Sprintf(" AND d.location ILIKE $%d", paramIndex)
		args = append(args, "%"+params.Location+"%")
		paramIndex++
	}

	// Work mode filtreleme
	if params.WorkMode != "" {
		whereClause += fmt.Sprintf(" AND d.work_mode = $%d", paramIndex)
		args = append(args, params.WorkMode)
		paramIndex++
	}

	// Arama sorgusu (başlık ve açıklamada)
	if params.Query != "" {
		whereClause += fmt.Sprintf(" AND (d.title ILIKE $%d OR d.description ILIKE $%d)", paramIndex, paramIndex+1)
		searchTerm := "%" + params.Query + "%"
		args = append(args, searchTerm, searchTerm)
		paramIndex += 2
	}

	// Sıralama
	var orderBy string
	switch params.SortBy {
	case "title":
		orderBy = "d.title"
	case "deadline":
		orderBy = "p.deadline"
	case "createdAt", "created_at":
		orderBy = "p.created_at"
	case "updatedAt", "updated_at":
		orderBy = "p.updated_at"
	default:
		orderBy = "p.created_at"
	}

	orderClause := fmt.Sprintf(" ORDER BY %s %s", orderBy, strings.ToUpper(params.SortOrder))
	limitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", params.Limit, (params.Page-1)*params.Limit)

	// Toplam sayıyı al
	var total int
	err := r.db.QueryRowContext(ctx, countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("iş ilanları sayısı alınamadı: %w", err)
	}

	// İş ilanlarını al
	rows, err := r.db.QueryContext(ctx, baseQuery+whereClause+orderClause+limitOffset, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("iş ilanları getirilemedi: %w", err)
	}
	defer rows.Close()

	// Sonuçları tara
	jobs, err := scanJobs(rows)
	if err != nil {
		return nil, 0, err
	}

	return jobs, total, nil
}

// GetJobBySlug - URL yapısına (slug) göre iş ilanını view olarak getirir
func (r *Repository) GetJobBySlug(ctx context.Context, slug string) (types.JobView, error) {
	defer utils.TimeTrack(time.Now(), "Job -> Get Job By Slug")
	query := jobBaseQuery + `
		WHERE p.slug = $1
			AND p.status = 'published'
			AND p.status != 'deleted'
	`

	row := r.db.QueryRowContext(ctx, query, slug)
	return scanJob(row)
}

// GetJobByID - ID'ye göre iş ilanını view olarak getirir
func (r *Repository) GetJobByID(ctx context.Context, id uuid.UUID) (types.JobView, error) {
	defer utils.TimeTrack(time.Now(), "Job -> Get Job By ID")

	query := jobBaseQuery + `
								WHERE p.id = $1 AND p.status != 'deleted'
				`

	row := r.db.QueryRowContext(ctx, query, id)
	return scanJob(row)
}
