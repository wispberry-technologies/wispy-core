-- SQL schema for pages table
CREATE TABLE IF NOT EXISTS pages (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    slug TEXT NOT NULL,
    description TEXT,
    author TEXT,
    layout TEXT,
    is_draft INTEGER DEFAULT 0,
    is_static INTEGER DEFAULT 1,
    require_auth INTEGER DEFAULT 0,
    required_roles TEXT,
    file_path TEXT,
    protected TEXT,
    custom_data TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    published_at DATETIME,
    site_domain TEXT,
    site_name TEXT
);

-- SQL schema for page_content table (language-specific and content data)
CREATE TABLE IF NOT EXISTS page_content (
    id TEXT PRIMARY KEY,
    page_id TEXT NOT NULL,
    lang TEXT NOT NULL DEFAULT 'en',
    keywords TEXT,
    meta_tags TEXT,
    content_json TEXT, -- JSONB in Postgres, TEXT for SQLite
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(page_id) REFERENCES pages(id) ON DELETE CASCADE
);
