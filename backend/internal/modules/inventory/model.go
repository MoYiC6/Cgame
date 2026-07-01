package inventory

import "time"

type GoodsCategory struct {
	ID        int64
	ParentID  int64
	Name      string
	Icon      *string
	Sort      int
	Status    *int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Goods struct {
	ID                int64
	CategoryID        *int64
	Platform          *string
	Name              string
	Description       *string
	CoverImage        *string
	BillingMode       *string
	Status            *int
	IsVisible         bool
	CommissionType    *string
	CommissionRate    *float64
	MinTeacherLevel   int
	MapSelectEnabled  bool
	Version           int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type GoodsSKU struct {
	ID          int64
	GoodsID     int64
	SKUName     string
	SKUSnapshot map[string]any
	Price       float64
	Stock       int
	Sort        int
	Status      *int
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GoodsSpec struct {
	ID          int64
	GoodsID     int64
	SpecName    string
	SpecValues  []string
	Sort        int
	CreatedAt   time.Time
}

type GoodsSKUStockLog struct {
	ID        int64
	SKUID     int64
	OldStock  *int
	NewStock  *int
	ChangeType *string
	OrderID   *int64
	OperatorID *int64
	CreatedAt time.Time
}

type PurchaseLimitRule struct {
	ID          int64
	GoodsID     *int64
	LimitType   string
	LimitCount  int
	LimitPeriod int
	StartTime   *time.Time
	EndTime     *time.Time
	Status      *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
