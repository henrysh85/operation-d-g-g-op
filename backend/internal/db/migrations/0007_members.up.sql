-- 0007_members.up.sql: trade-association membership pipeline
CREATE TABLE IF NOT EXISTS members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    legal_name      TEXT NOT NULL,
    jurisdiction_id UUID REFERENCES countries(id) ON DELETE SET NULL,
    status          TEXT NOT NULL DEFAULT 'prospect',  -- prospect | applied | active | lapsed
    tier            TEXT,                                -- standard | premium | enterprise
    contact_id      UUID REFERENCES contacts(id) ON DELETE SET NULL,
    risk_score      INT,
    joined_at       DATE,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_members_status ON members(status);
CREATE INDEX IF NOT EXISTS idx_members_jurisdiction ON members(jurisdiction_id);
CREATE INDEX IF NOT EXISTS idx_members_name_trgm ON members USING gin (legal_name gin_trgm_ops);
