-- 0005_clients_activities.up.sql
CREATE TABLE IF NOT EXISTS clients (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        TEXT UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    kind        TEXT,
    tier        TEXT,
    status      TEXT NOT NULL DEFAULT 'active',
    metrics     JSONB NOT NULL DEFAULT '{}'::jsonb,
    cross_links JSONB NOT NULL DEFAULT '[]'::jsonb,
    verticals   TEXT[] NOT NULL DEFAULT '{}',
    tags        TEXT[] NOT NULL DEFAULT '{}',
    onboarded_at DATE,
    metadata    JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_clients_verticals ON clients USING gin (verticals);
CREATE INDEX IF NOT EXISTS idx_clients_tags ON clients USING gin (tags);

CREATE TABLE IF NOT EXISTS activities (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title      TEXT NOT NULL,
    description TEXT,
    type       TEXT NOT NULL,
    vertical   TEXT,
    region_id  UUID REFERENCES regions(id) ON DELETE SET NULL,
    country_id UUID REFERENCES countries(id) ON DELETE SET NULL,
    occurred_on DATE NOT NULL,
    impact     INT,
    owner_id   UUID REFERENCES people(id) ON DELETE SET NULL,
    status     TEXT NOT NULL DEFAULT 'done',
    highlight  BOOLEAN NOT NULL DEFAULT FALSE,
    metadata   JSONB NOT NULL DEFAULT '{}'::jsonb,
    search_tsv tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(title,'')), 'A') ||
        setweight(to_tsvector('english', coalesce(description,'')), 'B')
    ) STORED,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_activities_date ON activities(occurred_on DESC);
CREATE INDEX IF NOT EXISTS idx_activities_region ON activities(region_id);
CREATE INDEX IF NOT EXISTS idx_activities_country ON activities(country_id);
CREATE INDEX IF NOT EXISTS idx_activities_vertical ON activities(vertical);
CREATE INDEX IF NOT EXISTS idx_activities_owner ON activities(owner_id);
CREATE INDEX IF NOT EXISTS idx_activities_search ON activities USING gin (search_tsv);

CREATE TABLE IF NOT EXISTS activity_clients (
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    client_id   UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    PRIMARY KEY (activity_id, client_id)
);

CREATE TABLE IF NOT EXISTS activity_outputs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    label       TEXT NOT NULL,
    minio_key   TEXT NOT NULL,
    content_type TEXT,
    size_bytes  BIGINT,
    uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_activity_outputs_activity ON activity_outputs(activity_id);
