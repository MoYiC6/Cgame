package file

import (
	"context"
	"fmt"

	"backend/internal/platform/database"
)

type Repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) *Repository {
	return &Repository{dbtx: dbtx}
}

func (r *Repository) CreateCategory(ctx context.Context, c *FileCategory) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO file_categories (name, description, sort) VALUES ($1, $2, $3) RETURNING id`,
		c.Name, c.Description, c.Sort,
	).Scan(&c.ID)
}

func (r *Repository) GetCategory(ctx context.Context, id int64) (*FileCategory, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, name, description, sort, created_at, updated_at FROM file_categories WHERE id = $1`,
		id,
	)
	var c FileCategory
	err := row.Scan(&c.ID, &c.Name, &c.Description, &c.Sort, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get category: %w", err)
	}
	return &c, nil
}

func (r *Repository) ListCategories(ctx context.Context) ([]*FileCategory, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, name, description, sort, created_at, updated_at FROM file_categories ORDER BY sort ASC, id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var categories []*FileCategory
	for rows.Next() {
		var c FileCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Sort, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, &c)
	}
	return categories, nil
}

func (r *Repository) UpdateCategory(ctx context.Context, c *FileCategory) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE file_categories SET name = $1, description = $2, sort = $3, updated_at = NOW() WHERE id = $4`,
		c.Name, c.Description, c.Sort, c.ID,
	)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	return nil
}

func (r *Repository) DeleteCategory(ctx context.Context, id int64) error {
	_, err := r.dbtx.ExecContext(ctx, `DELETE FROM file_categories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	return nil
}

func (r *Repository) CreateFile(ctx context.Context, f *File) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO files (user_id, category_id, display_name, original_name, url, file_id, file_hash, type, size, provider, status, description, sort)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`,
		f.UserID, f.CategoryID, f.DisplayName, f.OriginalName, f.URL, f.FileID, f.FileHash, f.Type, f.Size, f.Provider, f.Status, f.Description, f.Sort,
	).Scan(&f.ID)
}

func (r *Repository) GetFile(ctx context.Context, id int64) (*File, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, user_id, category_id, display_name, original_name, url, file_id, file_hash, type, size, provider, status, description, sort, created_at, updated_at, deleted
		 FROM files WHERE id = $1 AND deleted = 0`,
		id,
	)
	var f File
	err := row.Scan(&f.ID, &f.UserID, &f.CategoryID, &f.DisplayName, &f.OriginalName, &f.URL, &f.FileID, &f.FileHash, &f.Type, &f.Size, &f.Provider, &f.Status, &f.Description, &f.Sort, &f.CreatedAt, &f.UpdatedAt, &f.Deleted)
	if err != nil {
		return nil, fmt.Errorf("get file: %w", err)
	}
	return &f, nil
}

func (r *Repository) ListFiles(ctx context.Context, userID *int64, categoryID *int64, fileType *string, keyword string, page, pageSize int) ([]*File, int, error) {
	where := "WHERE deleted = 0"
	args := []interface{}{}
	idx := 1

	if userID != nil {
		where += fmt.Sprintf(" AND user_id = $%d", idx)
		args = append(args, *userID)
		idx++
	}
	if categoryID != nil {
		where += fmt.Sprintf(" AND category_id = $%d", idx)
		args = append(args, *categoryID)
		idx++
	}
	if fileType != nil && *fileType != "" {
		where += fmt.Sprintf(" AND type = $%d", idx)
		args = append(args, *fileType)
		idx++
	}
	if keyword != "" {
		where += fmt.Sprintf(" AND (original_name ILIKE $%d OR display_name ILIKE $%d)", idx, idx+1)
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
		idx += 2
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM files %s", where)
	var total int
	if err := r.dbtx.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count files: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT id, user_id, category_id, display_name, original_name, url, file_id, file_hash, type, size, provider, status, description, sort, created_at, updated_at, deleted
		 FROM files %s ORDER BY id DESC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1,
	)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.dbtx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list files: %w", err)
	}
	defer rows.Close()

	var files []*File
	for rows.Next() {
		var f File
		if err := rows.Scan(&f.ID, &f.UserID, &f.CategoryID, &f.DisplayName, &f.OriginalName, &f.URL, &f.FileID, &f.FileHash, &f.Type, &f.Size, &f.Provider, &f.Status, &f.Description, &f.Sort, &f.CreatedAt, &f.UpdatedAt, &f.Deleted); err != nil {
			return nil, 0, fmt.Errorf("scan file: %w", err)
		}
		files = append(files, &f)
	}
	return files, total, nil
}

func (r *Repository) UpdateFile(ctx context.Context, f *File) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE files SET display_name = $1, category_id = $2, description = $3, sort = $4, updated_at = NOW() WHERE id = $5 AND deleted = 0`,
		f.DisplayName, f.CategoryID, f.Description, f.Sort, f.ID,
	)
	if err != nil {
		return fmt.Errorf("update file: %w", err)
	}
	return nil
}

func (r *Repository) DeleteFile(ctx context.Context, id int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE files SET deleted = 1, updated_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

func (r *Repository) FindFileByHash(ctx context.Context, hash string) (*File, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, user_id, category_id, display_name, original_name, url, file_id, file_hash, type, size, provider, status, description, sort, created_at, updated_at, deleted
		 FROM files WHERE file_hash = $1 AND deleted = 0 LIMIT 1`,
		hash,
	)
	var f File
	err := row.Scan(&f.ID, &f.UserID, &f.CategoryID, &f.DisplayName, &f.OriginalName, &f.URL, &f.FileID, &f.FileHash, &f.Type, &f.Size, &f.Provider, &f.Status, &f.Description, &f.Sort, &f.CreatedAt, &f.UpdatedAt, &f.Deleted)
	if err != nil {
		return nil, fmt.Errorf("find file by hash: %w", err)
	}
	return &f, nil
}
