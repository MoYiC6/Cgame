-- +goose Up
CREAT TABLE platform_runtime_probes_broken (
    id BIGSERIAL PRIMARY KEY
);

-- +goose Down
DROP TABLE IF EXISTS platform_runtime_probes_broken;
