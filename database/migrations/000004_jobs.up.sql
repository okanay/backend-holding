-- İş İlanı Durumu ENUM'u
CREATE TYPE job_status AS ENUM ('draft', 'published', 'closed', 'deleted');

-- İş İlanları Ana Tablosu
CREATE TABLE IF NOT EXISTS job_postings (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id),
    slug TEXT NOT NULL UNIQUE,
    status job_status DEFAULT 'draft' NOT NULL,
    deadline TIMESTAMPTZ, -- Başvuru son tarihi
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL
);

-- İş İlanı Detayları Tablosu
CREATE TABLE IF NOT EXISTS job_posting_details (
    id UUID NOT NULL REFERENCES job_postings (id) ON DELETE CASCADE PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    image TEXT,
    location TEXT, -- İş lokasyonu (İstanbul, Remote, vb.)
    work_mode TEXT, -- Çalışma şekli (Remote, On-site, Hybrid, vb.)
    employment_type TEXT, -- Çalışma tipi (Tam zamanlı, Yarı zamanlı, Staj, vb.)
    experience_level TEXT, -- Deneyim seviyesi (Junior, Mid-level, Senior)
    html TEXT NOT NULL, -- React Tiptap Editor HTML
    json TEXT NOT NULL, -- React Tiptap Editor JSON
    form_type TEXT NOT NULL, -- Form tipi ('basic', 'developer', 'designer', vs.)
    applicants INTEGER DEFAULT 0
);

-- İş Kategorileri Tablosu
CREATE TABLE IF NOT EXISTS job_categories (
    name TEXT NOT NULL PRIMARY KEY, -- "backend-developer"
    display_name TEXT NOT NULL, -- "Backend Developer"
    user_id UUID NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL
);

-- İlan-Kategori İlişki Tablosu
CREATE TABLE IF NOT EXISTS job_posting_categories (
    job_id UUID REFERENCES job_postings (id) ON DELETE CASCADE,
    category_name TEXT REFERENCES job_categories (name) ON DELETE CASCADE,
    PRIMARY KEY (job_id, category_name)
);

-- Başvurular Tablosu
CREATE TABLE IF NOT EXISTS job_applications (
    id UUID DEFAULT uuid_generate_v4 () PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES job_postings (id) ON DELETE CASCADE,
    full_name TEXT NOT NULL, -- Başvuran kişinin tam adı
    email TEXT NOT NULL, -- İletişim e-postası
    phone TEXT NOT NULL, -- İletişim telefonu (opsiyonel)
    form_type TEXT NOT NULL, -- Form tipi ('basic', 'developer', 'designer', vs.)
    form_json JSONB NOT NULL, -- Form verileri JSON formatında
    status TEXT NOT NULL DEFAULT 'received', -- Serbest metin
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL
);

-- İndeksler
CREATE INDEX idx_job_postings_user_id ON job_postings (user_id);

CREATE INDEX idx_job_categories_user_id ON job_categories (user_id);

CREATE INDEX idx_job_posting_categories_category_name ON job_posting_categories (category_name);

CREATE INDEX idx_job_applications_job_id ON job_applications (job_id);

CREATE INDEX idx_job_applications_status ON job_applications (status);

-- INSERT INTO job_categories (name, display_name, user_id) VALUES
--     ('software', 'Software Development', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('it', 'IT & Support', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('design', 'Design & UX', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('product', 'Product Management', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('data', 'Data & Analytics', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('marketing', 'Marketing', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('sales', 'Sales', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('customer-service', 'Customer Service', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('finance', 'Finance & Accounting', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('hr', 'Human Resources', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('operations', 'Operations', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('legal', 'Legal', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('admin', 'Administration', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('engineering', 'Engineering', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('education', 'Education', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('healthcare', 'Healthcare', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('logistics', 'Logistics', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('construction', 'Construction', 'f69d7689-8322-47fd-a438-116528b8c95e'),
--     ('other', 'Other', 'f69d7689-8322-47fd-a438-116528b8c95e');
