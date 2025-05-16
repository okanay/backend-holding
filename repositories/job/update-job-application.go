package JobRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-holding/utils"
)

func (r *Repository) UpdateJobApplicationStatus(ctx context.Context, applicationID uuid.UUID, status string) error {
	defer utils.TimeTrack(time.Now(), "Job -> Update Job Application Status")

	query := `
		UPDATE job_applications
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, status, applicationID)
	if err != nil {
		return fmt.Errorf("başvuru durumu güncellenemedi: %w", err)
	}

	return nil
}
