-- Bir cleanup fonksiyonu
CREATE OR REPLACE FUNCTION cleanup_tracking_records()
RETURNS void AS $$
BEGIN
  -- 30 günden eski kullanılmış kodları sil
  DELETE FROM jobs_application_tracking_codes
  WHERE (expires_at < NOW() OR is_used = TRUE)
  AND created_at < NOW() - INTERVAL '30 days';

  -- Süresi dolmuş oturumları sil
  DELETE FROM jobs_application_tracking_sessions
  WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER jobs_application_tracking_sessions_updated_at
BEFORE UPDATE ON jobs_application_tracking_sessions
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER jobs_application_tracking_codes_updated_at
BEFORE UPDATE ON jobs_application_tracking_codes
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
