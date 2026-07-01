-- name: GetOAuthByUserIDAndPlatform :one
SELECT id, user_id, platform, open_id, union_id, nickname, avatar, session_key, phone, bound_at, created_at, updated_at
FROM user_oauth
WHERE user_id = $1 AND platform = $2;

-- name: GetOAuthByOpenID :one
SELECT id, user_id, platform, open_id, union_id, nickname, avatar, session_key, phone, bound_at, created_at, updated_at
FROM user_oauth
WHERE platform = $1 AND open_id = $2;

-- name: CreateOAuth :one
INSERT INTO user_oauth (user_id, platform, open_id, union_id, nickname, avatar, session_key, phone, bound_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
RETURNING id;

-- name: GetUserTokenByUserID :one
SELECT id, user_id, access_token, refresh_token, expires_at, created_at
FROM user_tokens
WHERE user_id = $1;

-- name: CreateUserToken :one
INSERT INTO user_tokens (user_id, access_token, refresh_token, expires_at, created_at)
VALUES ($1, $2, $3, $4, NOW())
RETURNING id;

-- name: GetScanLoginSession :one
SELECT login_key, status, user_id, token, expires_at, created_at, updated_at
FROM scan_login_sessions
WHERE login_key = $1;

-- name: CreateScanLoginSession :one
INSERT INTO scan_login_sessions (login_key, status, user_id, token, expires_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING login_key;

-- name: UpdateScanLoginSession :exec
UPDATE scan_login_sessions
SET status = $2, user_id = $3, token = $4, updated_at = NOW()
WHERE login_key = $1;
