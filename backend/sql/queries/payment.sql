-- name: CreatePayment :one
INSERT INTO payment_records (payment_no, order_no, user_id, amount, status, pay_method, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING id;

-- name: GetPaymentByNo :one
SELECT id, payment_no, order_no, user_id, amount, status, pay_method, paid_at, refund_at, created_at, updated_at
FROM payment_records
WHERE payment_no = $1;

-- name: GetPaymentByOrderNo :one
SELECT id, payment_no, order_no, user_id, amount, status, pay_method, paid_at, refund_at, created_at, updated_at
FROM payment_records
WHERE order_no = $1;

-- name: ListPayments :many
SELECT id, payment_no, order_no, user_id, amount, status, pay_method, paid_at, refund_at, created_at, updated_at
FROM payment_records
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPayments :one
SELECT COUNT(*) FROM payment_records
WHERE user_id = $1;

-- name: UpdatePaymentStatus :exec
UPDATE payment_records
SET status = $2, paid_at = $3, refund_at = $4, updated_at = NOW()
WHERE payment_no = $1;
