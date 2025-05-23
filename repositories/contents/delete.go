package ContentRepository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

func (r *Repository) SoftDeleteContent(ctx context.Context, contentID uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Repository -> SoftDeleteContent")

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("tx begin error: %w", err)
	}
	defer tx.Rollback()

	var originalSlug string
	err = tx.QueryRowContext(ctx, "SELECT slug FROM contents WHERE id = $1 AND status != $2", contentID, types.ContentStatusDeleted).Scan(&originalSlug)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("soft delete için içerik bulunamadı veya zaten silinmiş (ID: %s)", contentID)
		}
		return fmt.Errorf("orijinal slug alınamadı: %w", err)
	}

	randomSuffix := utils.GenerateRandomString(8)
	newSlug := fmt.Sprintf("%s-DELETED-%s", originalSlug, randomSuffix)

	query := `UPDATE contents SET status = $1, slug = $2, updated_at = NOW() WHERE id = $3 AND status != $1`
	result, err := tx.ExecContext(ctx, query, types.ContentStatusDeleted, newSlug, contentID)
	if err != nil {
		return fmt.Errorf("Repository.SoftDeleteContent: exec error: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Repository.SoftDeleteContent: rows affected error: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("soft delete için uygun içerik bulunamadı (ID: %s)", contentID)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("tx commit error: %w", err)
	}
	return nil
}

func (r *Repository) HardDeleteContent(ctx context.Context, contentID uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Repository -> HardDeleteContent")

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("tx begin error: %w", err)
	}
	defer tx.Rollback()

	query := "DELETE FROM contents WHERE id = $1"
	result, err := tx.ExecContext(ctx, query, contentID)
	if err != nil {
		return fmt.Errorf("Repository.HardDeleteContent: exec error: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Repository.HardDeleteContent: rows affected error: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("hard delete için içerik bulunamadı (ID: %s)", contentID)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("tx commit error: %w", err)
	}
	return nil
}
