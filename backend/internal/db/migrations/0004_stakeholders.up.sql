-- 0004_stakeholders.up.sql
CREATE TABLE IF NOT EXISTS institutions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    kind        TEXT,
    country_id  UUID REFERENCES countries(id) ON DELETE SET NULL,
    website     TEXT,
    tags        TEXT[] NOT NULL DEFAULT '{}',
    metadata    JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_institutions_country ON institutions(country_id);
CREATE INDEX IF NOT EXISTS idx_institutions_name_trgm ON institutions USING gin (name gin_trgm_ops);

CREATE TABLE IF NOT EXISTS contact_groups (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS contacts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name           TEXT NOT NULL,
    email          TEXT,
    phone          TEXT,
    title          TEXT,
    institution_id UUID REFERENCES institutions(id) ON DELETE SET NULL,
    group_id       UUID REFERENCES contact_groups(id) ON DELETE SET NULL,
    relationship   INT CHECK (relationship BETWEEN 1 AND 5),
    dcgg_flag      BOOLEAN NOT NULL DEFAULT FALSE,
    verticals      TEXT[] NOT NULL DEFAULT '{}',
    tags           TEXT[] NOT NULL DEFAULT '{}',
    notes          TEXT,
    metadata       JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_contacts_institution ON contacts(institution_id);
CREATE INDEX IF NOT EXISTS idx_contacts_group ON contacts(group_id);
CREATE INDEX IF NOT EXISTS idx_contacts_verticals ON contacts USING gin (verticals);
CREATE INDEX IF NOT EXISTS idx_contacts_tags ON contacts USING gin (tags);
CREATE INDEX IF NOT EXISTS idx_contacts_name_trgm ON contacts USING gin (name gin_trgm_ops);
