-- İndeksleri kaldır
DROP INDEX IF EXISTS idx_jobs_application_tracking_sessions_token;

DROP INDEX IF EXISTS idx_jobs_application_tracking_sessions_email;

DROP INDEX IF EXISTS idx_jobs_application_tracking_codes_tracking_code;

DROP INDEX IF EXISTS idx_jobs_application_tracking_codes_email;

-- Tabloları kaldır
DROP TABLE IF EXISTS jobs_application_tracking_sessions;

DROP TABLE IF EXISTS jobs_application_tracking_codes;
