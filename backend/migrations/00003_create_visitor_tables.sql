-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS visitor_sessions (
    id BIGSERIAL PRIMARY KEY,
    visitor_id VARCHAR(64) NOT NULL,
    session_id VARCHAR(64) NOT NULL UNIQUE,
    first_visit_time TIMESTAMPTZ NOT NULL,
    last_visit_time TIMESTAMPTZ NOT NULL,
    page_count INTEGER DEFAULT 1,
    platform VARCHAR(32),
    version VARCHAR(32),
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS visitor_page_views (
    id BIGSERIAL PRIMARY KEY,
    visitor_id VARCHAR(64) NOT NULL,
    session_id VARCHAR(64) NOT NULL,
    page_path VARCHAR(255) NOT NULL,
    page_title VARCHAR(255),
    visit_time TIMESTAMPTZ NOT NULL,
    duration INTEGER,
    referrer VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS visitor_daily_stats (
    id BIGSERIAL PRIMARY KEY,
    stat_date DATE NOT NULL UNIQUE,
    unique_visitors INTEGER,
    total_sessions INTEGER,
    total_page_views INTEGER,
    avg_session_duration NUMERIC(10,2),
    bounce_rate NUMERIC(10,2),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_visitor_sessions_visitor_id ON visitor_sessions(visitor_id);
CREATE INDEX IF NOT EXISTS idx_visitor_sessions_last_visit_time ON visitor_sessions(last_visit_time);
CREATE INDEX IF NOT EXISTS idx_visitor_page_views_session_id ON visitor_page_views(session_id);
CREATE INDEX IF NOT EXISTS idx_visitor_page_views_visit_time ON visitor_page_views(visit_time);
CREATE INDEX IF NOT EXISTS idx_visitor_daily_stats_stat_date ON visitor_daily_stats(stat_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS visitor_daily_stats;
DROP TABLE IF EXISTS visitor_page_views;
DROP TABLE IF EXISTS visitor_sessions;
-- +goose StatementEnd
