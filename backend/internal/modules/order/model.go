package order

type Order struct {
	OrderNo string
	UserID  int64
	SKU     string
	Qty     int
	Status  string
}
