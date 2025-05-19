package JobRepository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (r *Repository) ListJobsApplications(ctx context.Context, params types.JobApplicationSearchParams) ([]types.JobApplication, int, error) {
	defer utils.TimeTrack(time.Now(), "Job -> Get Job Applications")

	baseQuery := `
		SELECT
			a.id,
			a.job_id,
			a.full_name,
			a.email,
			a.phone,
			a.form_type,
			a.form_json,
			a.status,
			a.created_at,
			a.updated_at,
			d.title AS job_title
		FROM job_applications a
		LEFT JOIN job_postings p ON a.job_id = p.id
		LEFT JOIN job_posting_details d ON p.id = d.id
	`

	countQuery := `
		SELECT COUNT(*)
		FROM job_applications a
		LEFT JOIN job_postings p ON a.job_id = p.id
		LEFT JOIN job_posting_details d ON p.id = d.id
	`

	whereClause := " WHERE 1=1"
	args := []any{}
	paramIndex := 1

	if params.JobID != uuid.Nil {
		whereClause += fmt.Sprintf(" AND a.job_id = $%d", paramIndex)
		args = append(args, params.JobID)
		paramIndex++
	}

	if params.Status != "" {
		whereClause += fmt.Sprintf(" AND a.status = $%d", paramIndex)
		args = append(args, params.Status)
		paramIndex++
	}

	if params.FullName != "" {
		whereClause += fmt.Sprintf(" AND a.full_name = $%d", paramIndex)
		args = append(args, params.FullName)
		paramIndex++
	}

	if params.Email != "" {
		whereClause += fmt.Sprintf(" AND a.email = $%d", paramIndex)
		args = append(args, params.Email)
		paramIndex++
	}

	if params.StartDate != "" {
		whereClause += fmt.Sprintf(" AND a.created_at >= $%d", paramIndex)
		args = append(args, params.StartDate)
		paramIndex++
	}

	if params.EndDate != "" {
		whereClause += fmt.Sprintf(" AND a.created_at <= $%d", paramIndex)
		args = append(args, params.EndDate)
		paramIndex++
	}

	orderClause := fmt.Sprintf(" ORDER BY a.%s %s", params.SortBy, params.SortOrder)
	limitOffset := fmt.Sprintf(" LIMIT %d OFFSET %d", params.Limit, (params.Page-1)*params.Limit)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("başvuru sayısı alınamadı: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, baseQuery+whereClause+orderClause+limitOffset, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("başvurular getirilemedi: %w", err)
	}
	defer rows.Close()

	var applications []types.JobApplication
	for rows.Next() {
		var app types.JobApplication
		var jobTitle sql.NullString

		if err := rows.Scan(
			&app.ID,
			&app.JobID,
			&app.FullName,
			&app.Email,
			&app.Phone,
			&app.FormType,
			&app.FormJSON,
			&app.Status,
			&app.CreatedAt,
			&app.UpdatedAt,
			&jobTitle,
		); err != nil {
			return nil, 0, fmt.Errorf("başvuru bilgisi okunamadı: %w", err)
		}

		applications = append(applications, app)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("başvurular okunurken hata: %w", err)
	}

	return applications, total, nil
}

func (r *Repository) GetJobApplicationByID(ctx context.Context, applicationID uuid.UUID) (types.JobApplication, error) {
	defer utils.TimeTrack(time.Now(), "Job -> Get Job Application By ID")

	var app types.JobApplication
	var jobTitle sql.NullString

	query := `
		SELECT
			a.id,
			a.job_id,
			a.full_name,
			a.email,
			a.phone,
			a.form_type,
			a.form_json,
			a.status,
			a.created_at,
			a.updated_at,
			d.title AS job_title
		FROM job_applications a
		LEFT JOIN job_postings p ON a.job_id = p.id
		LEFT JOIN job_posting_details d ON p.id = d.id
		WHERE a.id = $1
	`

	err := r.db.QueryRowContext(ctx, query, applicationID).Scan(
		&app.ID,
		&app.JobID,
		&app.FullName,
		&app.Email,
		&app.Phone,
		&app.FormType,
		&app.FormJSON,
		&app.Status,
		&app.CreatedAt,
		&app.UpdatedAt,
		&jobTitle,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return app, nil
		}
		return app, fmt.Errorf("başvuru getirilemedi: %w", err)
	}

	return app, nil
}

func (r *Repository) GetJobApplicationsByEmail(ctx context.Context, email string) ([]types.JobApplication, error) {
	defer utils.TimeTrack(time.Now(), "Job -> Get Job Applications By Email")

	query := `
		SELECT
			a.id,
			a.job_id,
			a.full_name,
			a.email,
			a.phone,
			a.form_type,
			a.form_json,
			a.status,
			a.created_at,
			a.updated_at,
			d.title AS job_title
		FROM job_applications a
		LEFT JOIN job_postings p ON a.job_id = p.id
		LEFT JOIN job_posting_details d ON p.id = d.id
		WHERE a.email = $1
		ORDER BY a.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, email)
	if err != nil {
		return nil, fmt.Errorf("başvurular getirilemedi: %w", err)
	}
	defer rows.Close()

	var applications []types.JobApplication
	for rows.Next() {
		var app types.JobApplication
		var jobTitle sql.NullString

		if err := rows.Scan(
			&app.ID,
			&app.JobID,
			&app.FullName,
			&app.Email,
			&app.Phone,
			&app.FormType,
			&app.FormJSON,
			&app.Status,
			&app.CreatedAt,
			&app.UpdatedAt,
			&jobTitle,
		); err != nil {
			return nil, fmt.Errorf("başvuru bilgisi okunamadı: %w", err)
		}

		applications = append(applications, app)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("başvurular okunurken hata: %w", err)
	}

	return applications, nil
}
