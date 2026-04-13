-- 0006_content.up.sql
CREATE TABLE IF NOT EXISTS templates (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug       TEXT UNIQUE NOT NULL,
    name       TEXT NOT NULL,
    kind       TEXT NOT NULL,
    body       TEXT NOT NULL,
    variables  JSONB NOT NULL DEFAULT '[]'::jsonb,
    tags       TEXT[] NOT NULL DEFAULT '{}',
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_templates_kind ON templates(kind);

CREATE TABLE IF NOT EXISTS publications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title      TEXT NOT NULL,
    authors    TEXT[] NOT NULL DEFAULT '{}',
    venue      TEXT,
    url        TEXT,
    abstract   TEXT,
    published_on DATE,
    verticals  TEXT[] NOT NULL DEFAULT '{}',
    tags       TEXT[] NOT NULL DEFAULT '{}',
    metadata   JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_publications_published ON publications(published_on DESC);
CREATE INDEX IF NOT EXISTS idx_publications_verticals ON publications USING gin (verticals);

CREATE TABLE IF NOT EXISTS memberships_generated (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id  UUID REFERENCES templates(id) ON DELETE SET NULL,
    contact_id   UUID REFERENCES contacts(id) ON DELETE SET NULL,
    institution_id UUID REFERENCES institutions(id) ON DELETE SET NULL,
    rendered     TEXT NOT NULL,
    params       JSONB NOT NULL DEFAULT '{}'::jsonb,
    minio_key    TEXT,
    generated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_memberships_template ON memberships_generated(template_id);
CREATE INDEX IF NOT EXISTS idx_memberships_contact ON memberships_generated(contact_id);
