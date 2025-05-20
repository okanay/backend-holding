package JobRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-holding/types"
	"github.com/okanay/backend-holding/utils"
)

// DeleteJob iş ilanını siler (soft veya hard delete seçeneği ile)
func (r *Repository) DeleteJob(ctx context.Context, jobID uuid.UUID, hardDelete bool) error {
	defer utils.TimeTrack(time.Now(), "Job -> Delete Job")

	if hardDelete {
		return r.HardDeleteJob(ctx, jobID)
	} else {
		return r.SoftDeleteJob(ctx, jobID)
	}
}

// HardDeleteJob iş ilanını ve ilgili tüm kayıtları veritabanından tamamen siler
func (r *Repository) HardDeleteJob(ctx context.Context, jobID uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Job -> Hard Delete Job")

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("işlem başlatılamadı: %w", err)
	}
	defer tx.Rollback()

	// Önce kategori bağlantılarını sil
	_, err = tx.ExecContext(ctx, "DELETE FROM job_posting_categories WHERE job_id = $1", jobID)
	if err != nil {
		return fmt.Errorf("iş kategorileri silinemedi: %w", err)
	}

	// Başvuruları işaretle
	_, err = tx.ExecContext(ctx, "UPDATE job_applications SET status = 'job_deleted' WHERE job_id = $1", jobID)
	if err != nil {
		return fmt.Errorf("başvurular güncellenemedi: %w", err)
	}

	// İş detaylarını sil
	_, err = tx.ExecContext(ctx, "DELETE FROM job_posting_details WHERE id = $1", jobID)
	if err != nil {
		return fmt.Errorf("iş detayları silinemedi: %w", err)
	}

	// Son olarak ana iş ilanını sil
	result, err := tx.ExecContext(ctx, "DELETE FROM job_postings WHERE id = $1", jobID)
	if err != nil {
		return fmt.Errorf("iş ilanı silinemedi: %w", err)
	}

	// Etkilenen satır sayısını kontrol et
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("etkilenen satır sayısı alınamadı: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("silinecek iş ilanı bulunamadı")
	}

	// İşlemi commit et
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("işlem tamamlanamadı: %w", err)
	}

	return nil
}

// SoftDeleteJob iş ilanının durumunu "deleted" olarak işaretler ve slug'ı değiştirir
func (r *Repository) SoftDeleteJob(ctx context.Context, jobID uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Job -> Soft Delete Job")

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("işlem başlatılamadı: %w", err)
	}
	defer tx.Rollback()

	// Mevcut slug'ı al
	var originalSlug string
	if err := tx.QueryRowContext(ctx, "SELECT slug FROM job_postings WHERE id = $1", jobID).Scan(&originalSlug); err != nil {
		return fmt.Errorf("slug alınamadı: %w", err)
	}

	// Rastgele bir son ek oluştur
	randomSuffix := utils.GenerateRandomString(8)

	// Yeni slug oluştur: "original-slug-DELETED-RANDOMSTRING"
	newSlug := fmt.Sprintf("%s-DELETED-%s", originalSlug, randomSuffix)

	// Status ve slug'ı güncelle
	query := `
		UPDATE job_postings
		SET status = $1, slug = $2, updated_at = NOW()
		WHERE id = $3
	`

	result, err := tx.ExecContext(ctx, query, types.JobStatusDeleted, newSlug, jobID)
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

	// İşlemi commit et
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("işlem tamamlanamadı: %w", err)
	}

	return nil
}
