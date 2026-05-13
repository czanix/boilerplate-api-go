package domain

import "context"

type OrderRepository interface {
	Save(ctx context.Context, order *Order) error
	FindByPublicID(ctx context.Context, publicID string) (*Order, error)
	FindByCustomerID(ctx context.Context, customerID string) ([]*Order, error)
	Update(ctx context.Context, order *Order) error
	SoftDelete(ctx context.Context, publicID string) error
}
