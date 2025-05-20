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
--     ('software', 'Software Development', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('it', 'IT & Support', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('design', 'Design & UX', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('product', 'Product Management', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('data', 'Data & Analytics', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('marketing', 'Marketing', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('sales', 'Sales', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('customer-service', 'Customer Service', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('finance', 'Finance & Accounting', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('hr', 'Human Resources', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('operations', 'Operations', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('legal', 'Legal', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('admin', 'Administration', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('engineering', 'Engineering', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('education', 'Education', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('healthcare', 'Healthcare', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('logistics', 'Logistics', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('construction', 'Construction', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1'),
--     ('other', 'Other', 'be3a28e6-4be8-4bb9-8191-d37f3907f1b1');
