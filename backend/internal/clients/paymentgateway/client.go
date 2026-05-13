package paymentgateway

import "context"

type CreatePaymentRequest struct {
	OrderNo  string
	Amount   int64
	Currency string
}

type CreatePaymentResponse struct {
	PaymentNo string
	Status    string
}

type Client interface {
	CreatePayment(ctx context.Context, req CreatePaymentRequest) (*CreatePaymentResponse, error)
}
