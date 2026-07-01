-- name: CreateFile :one
INSERT INTO files (user_id, category_id, display_name, original_name, url, file_id, file_hash, type, size, provider, status, description, sort, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
RETURNING id;

-- name: GetFileByID :one
SELECT id, user_id, category_id, display_name, original_name, url, file_id, file_hash, type, size, provider, status, description, sort, created_at, updated_at
FROM files
WHERE id = $1;

-- name: GetFileByHash :one
SELECT id, user_id, category_id, display_name, original_name, url, file_id, file_hash, type, size, provider, status, description, sort, created_at, updated_at
FROM files
WHERE file_hash = $1;

-- name: ListFiles :many
SELECT id, user_id, category_id, display_name, original_name, url, file_id, file_hash, type, size, provider, status, description, sort, created_at, updated_at
FROM files
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountFiles :one
SELECT COUNT(*) FROM files;

-- name: UpdateFileStatus :exec
UPDATE files
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: SoftDeleteFile :exec
UPDATE files
SET deleted = 1, updated_at = NOW()
WHERE id = $1;
