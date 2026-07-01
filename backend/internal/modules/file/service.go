package file

import (
	"context"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateCategory(ctx context.Context, c *FileCategory) (int64, error) {
	if c.Name == "" {
		return 0, fmt.Errorf("category name is required")
	}
	if err := s.repo.CreateCategory(ctx, c); err != nil {
		return 0, err
	}
	return c.ID, nil
}

func (s *Service) GetCategory(ctx context.Context, id int64) (*FileCategory, error) {
	return s.repo.GetCategory(ctx, id)
}

func (s *Service) ListCategories(ctx context.Context) ([]*FileCategory, error) {
	return s.repo.ListCategories(ctx)
}

func (s *Service) UpdateCategory(ctx context.Context, c *FileCategory) error {
	if c.ID == 0 || c.Name == "" {
		return fmt.Errorf("invalid category")
	}
	return s.repo.UpdateCategory(ctx, c)
}

func (s *Service) DeleteCategory(ctx context.Context, id int64) error {
	if id <= 2 {
		return fmt.Errorf("cannot delete built-in category")
	}
	return s.repo.DeleteCategory(ctx, id)
}

func (s *Service) CreateFile(ctx context.Context, f *File) error {
	if f.URL == "" || f.OriginalName == nil || *f.OriginalName == "" {
		return fmt.Errorf("url and original name are required")
	}
	return s.repo.CreateFile(ctx, f)
}

func (s *Service) GetFile(ctx context.Context, id int64) (*File, error) {
	return s.repo.GetFile(ctx, id)
}

func (s *Service) ListFiles(ctx context.Context, userID *int64, categoryID *int64, fileType *string, keyword string, page, pageSize int) ([]*File, int, error) {
	return s.repo.ListFiles(ctx, userID, categoryID, fileType, keyword, page, pageSize)
}

func (s *Service) UpdateFile(ctx context.Context, f *File) error {
	if f.ID == 0 {
		return fmt.Errorf("file id is required")
	}
	return s.repo.UpdateFile(ctx, f)
}

func (s *Service) DeleteFile(ctx context.Context, id int64) error {
	return s.repo.DeleteFile(ctx, id)
}

func (s *Service) FindFileByHash(ctx context.Context, hash string) (*File, error) {
	return s.repo.FindFileByHash(ctx, hash)
}
