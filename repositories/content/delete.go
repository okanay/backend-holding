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

// SoftDeleteContent - İçeriği soft delete yapar (status'u deleted yapar ve slug'ı değiştirir)
func (r *Repository) SoftDeleteContent(ctx context.Context, contentID uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Repository -> SoftDeleteContent")

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("transaction başlatılamadı: %w", err)
	}
	defer tx.Rollback()

	// Mevcut slug'ı al
	var originalSlug string
	err = tx.QueryRowContext(ctx,
		"SELECT slug FROM contents WHERE id = $1 AND status != $2",
		contentID, types.ContentStatusDeleted).Scan(&originalSlug)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("içerik bulunamadı veya zaten silinmiş (ID: %s)", contentID)
		}
		return fmt.Errorf("slug alınamadı: %w", err)
	}

	// Benzersiz slug oluştur
	timestamp := time.Now().Unix()
	randomSuffix := utils.GenerateRandomString(6)
	newSlug := fmt.Sprintf("%s-DELETED-%d-%s", originalSlug, timestamp, randomSuffix)

	// Soft delete işlemi
	query := `
		UPDATE contents
		SET status = $1, slug = $2, updated_at = NOW()
		WHERE id = $3 AND status != $1
	`

	result, err := tx.ExecContext(ctx, query, types.ContentStatusDeleted, newSlug, contentID)
	if err != nil {
		return fmt.Errorf("soft delete hatası: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("etkilenen satır sayısı alınamadı: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("içerik zaten silinmiş (ID: %s)", contentID)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit hatası: %w", err)
	}

	return nil
}

// HardDeleteContent - İçeriği veritabanından kalıcı olarak siler
func (r *Repository) HardDeleteContent(ctx context.Context, contentID uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Repository -> HardDeleteContent")

	query := "DELETE FROM contents WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, contentID)
	if err != nil {
		return fmt.Errorf("hard delete hatası: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("etkilenen satır sayısı alınamadı: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("silinecek içerik bulunamadı (ID: %s)", contentID)
	}

	return nil
}

// RestoreContent - Soft delete edilmiş içeriği geri yükler
func (r *Repository) RestoreContent(ctx context.Context, contentID uuid.UUID, newSlug string) error {
	defer utils.TimeTrack(time.Now(), "Repository -> RestoreContent")

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("transaction başlatılamadı: %w", err)
	}
	defer tx.Rollback()

	// Yeni slug'ın müsait olup olmadığını kontrol et
	var count int
	err = tx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM contents WHERE slug = $1 AND id != $2 AND status != $3",
		newSlug, contentID, types.ContentStatusDeleted).Scan(&count)

	if err != nil {
		return fmt.Errorf("slug kontrolü yapılamadı: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("bu slug zaten kullanımda: %s", newSlug)
	}

	// Restore işlemi
	query := `
		UPDATE contents
		SET status = $1, slug = $2, updated_at = NOW()
		WHERE id = $3 AND status = $4
	`

	result, err := tx.ExecContext(ctx, query, types.ContentStatusDraft, newSlug, contentID, types.ContentStatusDeleted)
	if err != nil {
		return fmt.Errorf("restore hatası: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("etkilenen satır sayısı alınamadı: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("geri yüklenecek içerik bulunamadı (ID: %s)", contentID)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit hatası: %w", err)
	}

	return nil
}
