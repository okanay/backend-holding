-- İndeksleri kaldır
DROP INDEX IF EXISTS idx_contents_user_id;

DROP INDEX IF EXISTS idx_contents_language;

DROP INDEX IF EXISTS idx_contents_status;

DROP INDEX IF EXISTS idx_contents_slug;

DROP INDEX IF EXISTS idx_contents_identifier;

DROP INDEX IF EXISTS idx_contents_category;

DROP TABLE IF EXISTS contents;

DROP TYPE IF EXISTS content_status;
