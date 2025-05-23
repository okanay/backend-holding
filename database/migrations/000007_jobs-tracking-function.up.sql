-- Bir cleanup fonksiyonu
CREATE OR REPLACE FUNCTION cleanup_tracking_records()
RETURNS void AS $$
BEGIN
  -- 30 günden eski kullanılmış kodları sil
  DELETE FROM jobs_tracking_codes
  WHERE (expires_at < NOW() OR is_used = TRUE)
  AND created_at < NOW() - INTERVAL '30 days';

  -- Süresi dolmuş oturumları sil
  DELETE FROM jobs_tracking_sessions
  WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;
