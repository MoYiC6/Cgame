package payment

type PaymentOrder struct {
	PaymentNo string
	OrderNo   string
	Amount    int64
	Status    string
}
