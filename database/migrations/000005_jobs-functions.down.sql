-- Önce trigger'ları kaldırın
DROP TRIGGER IF EXISTS trg_applicants_after_insert ON job_applications;

DROP TRIGGER IF EXISTS trg_applicants_after_delete ON job_applications;

-- Sonra fonksiyonları kaldırın
DROP FUNCTION IF EXISTS update_applicants_count () CASCADE;

DROP FUNCTION IF EXISTS update_applicants_count_on_delete () CASCADE;
