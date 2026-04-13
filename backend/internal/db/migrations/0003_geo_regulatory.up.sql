-- 0003_geo_regulatory.up.sql
CREATE TABLE IF NOT EXISTS regions (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS countries (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code      TEXT UNIQUE NOT NULL,
    name      TEXT NOT NULL,
    region_id UUID REFERENCES regions(id) ON DELETE SET NULL,
    tier      INT,
    mena_sub  TEXT,
    metadata  JSONB NOT NULL DEFAULT '{}'::jsonb
);
CREATE INDEX IF NOT EXISTS idx_countries_region ON countries(region_id);
CREATE INDEX IF NOT EXISTS idx_countries_tier ON countries(tier);

CREATE TABLE IF NOT EXISTS jurisdictions_status (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    country_id     UUID NOT NULL REFERENCES countries(id) ON DELETE CASCADE,
    vertical       TEXT NOT NULL,
    status         TEXT NOT NULL,
    headline       TEXT,
    regulators     TEXT[] NOT NULL DEFAULT '{}',
    timeline       JSONB NOT NULL DEFAULT '[]'::jsonb,
    impact_matrix  JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_reviewed  TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (country_id, vertical)
);
CREATE INDEX IF NOT EXISTS idx_jurstat_vertical ON jurisdictions_status(vertical);
CREATE INDEX IF NOT EXISTS idx_jurstat_status ON jurisdictions_status(status);

CREATE TABLE IF NOT EXISTS consultations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jurisdiction_id UUID NOT NULL REFERENCES jurisdictions_status(id) ON DELETE CASCADE,
    vertical        TEXT NOT NULL,
    title           TEXT NOT NULL,
    deadline        DATE,
    status          TEXT NOT NULL DEFAULT 'open',
    assignee_id     UUID REFERENCES people(id) ON DELETE SET NULL,
    summary         TEXT,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_consult_jurisdiction ON consultations(jurisdiction_id);
CREATE INDEX IF NOT EXISTS idx_consult_vertical ON consultations(vertical);
CREATE INDEX IF NOT EXISTS idx_consult_deadline ON consultations(deadline);
CREATE INDEX IF NOT EXISTS idx_consult_assignee ON consultations(assignee_id);
