-- 0002_people.up.sql
CREATE TABLE IF NOT EXISTS people (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL,
    email         TEXT UNIQUE,
    dept          TEXT,
    title         TEXT,
    reports_to    UUID REFERENCES people(id) ON DELETE SET NULL,
    status        TEXT NOT NULL DEFAULT 'active',
    location      TEXT,
    country_code  TEXT,
    salary        NUMERIC(14,2),
    currency      TEXT DEFAULT 'USD',
    holiday_quota INT NOT NULL DEFAULT 25,
    photo_key     TEXT,
    start_date    DATE,
    metadata      JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_people_dept ON people(dept);
CREATE INDEX IF NOT EXISTS idx_people_reports_to ON people(reports_to);
CREATE INDEX IF NOT EXISTS idx_people_name_trgm ON people USING gin (name gin_trgm_ops);

CREATE TABLE IF NOT EXISTS holidays (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    person_id  UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date   DATE NOT NULL,
    days       NUMERIC(5,2) NOT NULL,
    status     TEXT NOT NULL DEFAULT 'pending',
    note       TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_holidays_person ON holidays(person_id);
CREATE INDEX IF NOT EXISTS idx_holidays_dates ON holidays(start_date, end_date);

CREATE TABLE IF NOT EXISTS reviews (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    person_id   UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    reviewer_id UUID REFERENCES people(id) ON DELETE SET NULL,
    period      TEXT NOT NULL,
    rating      INT,
    summary     TEXT,
    details     JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_reviews_person ON reviews(person_id);

CREATE TABLE IF NOT EXISTS expenses (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    person_id  UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    amount     NUMERIC(14,2) NOT NULL,
    currency   TEXT NOT NULL DEFAULT 'USD',
    category   TEXT,
    incurred_on DATE NOT NULL,
    memo       TEXT,
    receipt_key TEXT,
    status     TEXT NOT NULL DEFAULT 'submitted',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_expenses_person ON expenses(person_id);
CREATE INDEX IF NOT EXISTS idx_expenses_incurred ON expenses(incurred_on DESC);
