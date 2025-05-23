-- CONTENT STATUS ENUM'u
CREATE TYPE content_status AS ENUM ('draft', 'published', 'closed', 'deleted');

-- CONTENTS TABLOSU
CREATE TABLE IF NOT EXISTS contents (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    user_id UUID REFERENCES users (id) ON DELETE SET NULL,
    slug TEXT NOT NULL,
    identifier TEXT NOT NULL,
    language TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    category TEXT NOT NULL DEFAULT '',
    image_url TEXT,
    details_json JSONB,
    content_json JSONB NOT NULL,
    content_html TEXT NOT NULL,
    status content_status DEFAULT 'draft' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    CONSTRAINT uq_identifier_language UNIQUE (identifier, language),
    CONSTRAINT uq_slug_language UNIQUE (slug, language)
);

-- CONTENTS TABLOSU İÇİN İNDEKSLEr
CREATE INDEX IF NOT EXISTS idx_contents_identifier ON contents (identifier);

CREATE INDEX IF NOT EXISTS idx_contents_slug ON contents (slug);

CREATE INDEX IF NOT EXISTS idx_contents_status ON contents (status);

CREATE INDEX IF NOT EXISTS idx_contents_language ON contents (language);

CREATE INDEX IF NOT EXISTS idx_contents_user_id ON contents (user_id);

CREATE INDEX IF NOT EXISTS idx_contents_category ON contents (category);
