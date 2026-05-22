-- name: GetUserByEmail :one
SELECT id, public_id, email, password_hash, status, password_changed_at, last_login_at, created_at, updated_at
FROM users
WHERE email = $1;

-- name: ListUserRoles :many
SELECT r.code
FROM roles r
JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1
ORDER BY r.code;

-- name: ListUserPermissions :many
SELECT DISTINCT p.code
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
JOIN user_roles ur ON ur.role_id = rp.role_id
WHERE ur.user_id = $1
ORDER BY p.code;
