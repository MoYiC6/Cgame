-- +goose Up
CREATE TABLE platform_runtime_probes (
    id BIGSERIAL PRIMARY KEY,
    run_id TEXT NOT NULL,
    probe_name TEXT NOT NULL,
    observed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_platform_runtime_probes_run_id
    ON platform_runtime_probes (run_id);

-- +goose Down
DROP TABLE IF EXISTS platform_runtime_probes;
