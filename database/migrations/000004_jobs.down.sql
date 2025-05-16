-- Önce ilişki tabloları ve index'leri
DROP INDEX IF EXISTS idx_job_applications_status;

DROP INDEX IF EXISTS idx_job_applications_job_id;

DROP TABLE IF EXISTS job_applications;

DROP INDEX IF EXISTS idx_job_posting_categories_category_name;

DROP TABLE IF EXISTS job_posting_categories;

-- Bağımlı tablolar
DROP TABLE IF EXISTS job_posting_details;

DROP INDEX IF EXISTS idx_job_categories_user_id;

DROP TABLE IF EXISTS job_categories;

-- Ana tablo
DROP INDEX IF EXISTS idx_job_postings_user_id;

DROP TABLE IF EXISTS job_postings;

-- Enum tipi
DROP TYPE IF EXISTS job_status;
