package JobRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (r *Repository) CreateJobApplication(ctx context.Context, jobID uuid.UUID, input types.JobApplicationInput) (types.JobApplication, error) {
	defer utils.TimeTrack(time.Now(), "Job -> Create Job Application")
	var application types.JobApplication

	createQuery := `
		INSERT INTO job_applications (job_id, full_name, email, phone, form_type, form_json, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'received')
		RETURNING id, job_id, full_name, email, phone, form_type, form_json, status, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		createQuery,
		jobID,
		input.FullName,
		input.Email,
		input.Phone,
		input.FormType,
		input.FormJSON,
	).Scan(
		&application.ID,
		&application.JobID,
		&application.FullName,
		&application.Email,
		&application.Phone,
		&application.FormType,
		&application.FormJSON,
		&application.Status,
		&application.CreatedAt,
		&application.UpdatedAt,
	)

	if err != nil {
		return application, fmt.Errorf("başvuru oluşturulamadı: %w", err)
	}

	return application, nil
}
