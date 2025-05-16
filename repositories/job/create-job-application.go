package JobRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (r *Repository) CreateJobApplication(ctx context.Context, jobID uuid.UUID, input types.JobApplicationInput) (*types.JobApplication, error) {
	defer utils.TimeTrack(time.Now(), "Job -> Create Job Application")

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("işlem başlatılamadı: %w", err)
	}
	defer tx.Rollback()

	createQuery := `
		INSERT INTO job_applications (job_id, full_name, email, phone, form_type, form_json, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'received')
		RETURNING id, job_id, full_name, email, phone, form_type, form_json, status, created_at, updated_at
	`

	var application types.JobApplication
	err = tx.QueryRowContext(
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
		return nil, fmt.Errorf("başvuru oluşturulamadı: %w", err)
	}

	historyQuery := `
		INSERT INTO job_application_status_history (job_application_id, old_status, new_status)
		VALUES ($1, NULL, $2)
		RETURNING id, job_application_id, old_status, new_status, created_at, updated_at
	`

	var history types.JobApplicationStatusHistory
	err = tx.QueryRowContext(
		ctx,
		historyQuery,
		application.ID,
		application.Status,
	).Scan(
		&history.ID,
		&history.JobApplicationID,
		&history.OldStatus,
		&history.NewStatus,
		&history.CreatedAt,
		&history.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("başvuru durumu geçmişi oluşturulamadı: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("işlem tamamlanamadı: %w", err)
	}

	return &application, nil
}
