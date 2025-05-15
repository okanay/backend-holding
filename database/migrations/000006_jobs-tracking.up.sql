CREATE TABLE IF NOT EXISTS jobs_tracking_codes (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    email TEXT NOT NULL,
    tracking_code TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    is_used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW ()
);

-- İsteğe bağlı olarak bir de erişim oturumları tablosu ekleyebiliriz
CREATE TABLE IF NOT EXISTS jobs_tracking_sessions (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    email TEXT NOT NULL,
    session_token TEXT NOT NULL UNIQUE,
    ip_address TEXT,
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW ()
);

CREATE INDEX idx_jobs_tracking_codes_email ON jobs_tracking_codes (email);

CREATE INDEX idx_jobs_tracking_codes_tracking_code ON jobs_tracking_codes (tracking_code);

CREATE INDEX idx_jobs_tracking_sessions_email ON jobs_tracking_sessions (email);

CREATE INDEX idx_jobs_tracking_sessions_token ON jobs_tracking_sessions (session_token);
