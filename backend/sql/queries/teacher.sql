-- name: GetByID :one
SELECT id, user_id, name, mobile, avatar, status, rating, order_count, deposit, balance, platforms, tags, goods_ids, auto_status_enabled, created_at, updated_at
FROM teachers
WHERE id = $1;

-- name: GetByUserID :one
SELECT id, user_id, name, mobile, avatar, status, rating, order_count, deposit, balance, platforms, tags, goods_ids, auto_status_enabled, created_at, updated_at
FROM teachers
WHERE user_id = $1;

-- name: ListTeachers :many
SELECT id, user_id, name, mobile, avatar, status, rating, order_count, deposit, balance, platforms, tags, goods_ids, auto_status_enabled, created_at, updated_at
FROM teachers
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountTeachers :one
SELECT COUNT(*) FROM teachers;

-- name: CreateTeacher :one
INSERT INTO teachers (user_id, name, mobile, avatar, status, rating, order_count, deposit, balance, platforms, tags, goods_ids, auto_status_enabled, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
RETURNING id;

-- name: UpdateTeacherStatus :exec
UPDATE teachers
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: ListTeacherLevels :many
SELECT id, name, min_orders, commission_rate, priority, status, description, created_at, updated_at
FROM teacher_levels
ORDER BY priority DESC, id ASC;

-- name: CreateTeacherLevel :one
INSERT INTO teacher_levels (name, min_orders, commission_rate, priority, status, description, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING id;

-- name: UpdateTeacherLevel :exec
UPDATE teacher_levels
SET name = $2, min_orders = $3, commission_rate = $4, priority = $5, status = $6, description = $7, updated_at = NOW()
WHERE id = $1;

-- name: CreateTeacherStatusLog :one
INSERT INTO teacher_status_logs (teacher_id, old_status, new_status, reason, operator_id, created_at)
VALUES ($1, $2, $3, $4, $5, NOW())
RETURNING id;

-- name: GetTeacherStatusLogs :many
SELECT id, teacher_id, old_status, new_status, reason, operator_id, created_at
FROM teacher_status_logs
WHERE teacher_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
