-- name: GetSettingByKey :one
SELECT id, setting_key, setting_value, setting_type, category, description, is_public, created_at, updated_at
FROM system_settings
WHERE setting_key = $1;

-- name: ListSettingsByPrefix :many
SELECT id, setting_key, setting_value, setting_type, category, description, is_public, created_at, updated_at
FROM system_settings
WHERE setting_key LIKE $1 || '%'
ORDER BY setting_key ASC;

-- name: UpsertSetting :one
INSERT INTO system_settings (setting_key, setting_value, setting_type, category, description, is_public, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
ON CONFLICT (setting_key) DO UPDATE SET
    setting_value = EXCLUDED.setting_value,
    setting_type = EXCLUDED.setting_type,
    category = EXCLUDED.category,
    description = EXCLUDED.description,
    is_public = EXCLUDED.is_public,
    updated_at = NOW()
RETURNING id;

-- name: ListPartnerConfigs :many
SELECT id, config_key, config_value, description, status, created_at, updated_at
FROM partner_config
ORDER BY id ASC;

-- name: GetFaceidConfig :one
SELECT id, secret_id, secret_key, rule_id, region, redirect_url, is_enabled, remark, manual_enabled, created_at, updated_at
FROM faceid_config
WHERE deleted = 0
ORDER BY id ASC
LIMIT 1;

-- name: ListFaceidConfigs :many
SELECT id, secret_id, secret_key, rule_id, region, redirect_url, is_enabled, remark, manual_enabled, created_at, updated_at
FROM faceid_config
WHERE deleted = 0
ORDER BY id ASC;

-- name: UpdateFaceidConfig :exec
UPDATE faceid_config
SET secret_id = $2, secret_key = $3, rule_id = $4, region = $5, redirect_url = $6, is_enabled = $7, remark = $8, manual_enabled = $9, updated_at = NOW()
WHERE id = $1;
