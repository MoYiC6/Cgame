package user

import (
	"context"
	"database/sql"

	"backend/internal/platform/database"
)

type Repository interface {
	GetByID(ctx context.Context, userID int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) Repository {
	return &repository{dbtx: dbtx}
}

func (r *repository) GetByID(ctx context.Context, userID int64) (*User, error) {
	row := r.executor(ctx).QueryRowContext(ctx, `
		SELECT id, public_id, email, password_hash, status, password_changed_at, last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`, userID)
	return scanUser(row)
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := r.executor(ctx).QueryRowContext(ctx, `
		SELECT id, public_id, email, password_hash, status, password_changed_at, last_login_at, created_at, updated_at
		FROM users
		WHERE email = $1
	`, NormalizeEmail(email))
	return scanUser(row)
}

func (r *repository) executor(ctx context.Context) database.DBTX {
	return database.ExecutorFromContext(ctx, r.dbtx)
}

func scanUser(row rowScanner) (*User, error) {
	if row == nil {
		return nil, nil
	}
	var user User
	err := row.Scan(
		&user.ID,
		&user.PublicID,
		&user.Email,
		&user.PasswordHash,
		&user.Status,
		&user.PasswordChangedAt,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}
