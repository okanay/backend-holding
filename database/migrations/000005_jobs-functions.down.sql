-- Önce trigger'ları kaldırın
DROP TRIGGER IF EXISTS update_job_postings_updated_at ON job_postings;

DROP TRIGGER IF EXISTS update_job_metadata_updated_at ON job_metadata;

DROP TRIGGER IF EXISTS update_job_categories_updated_at ON job_categories;

DROP TRIGGER IF EXISTS update_job_applications_updated_at ON job_applications;

DROP TRIGGER IF EXISTS update_job_application_status_history_updated_at ON job_application_status_history;

DROP TRIGGER IF EXISTS trg_applicants_after_insert ON job_applications;

DROP TRIGGER IF EXISTS trg_applicants_after_delete ON job_applications;

-- Sonra fonksiyonları kaldırın
DROP FUNCTION IF EXISTS update_applicants_count () CASCADE;

DROP FUNCTION IF EXISTS update_applicants_count_on_delete () CASCADE;
