package file

import "time"

type FileCategory struct {
	ID          int64
	Name        string
	Description *string
	Sort        int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type File struct {
	ID          int64
	UserID      *int64
	CategoryID  *int64
	DisplayName *string
	OriginalName *string
	URL         string
	FileID      *string
	FileHash    *string
	Type        *string
	Size        *int64
	Provider    *string
	Status      *int
	Description *string
	Sort        *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Deleted     *int
}
