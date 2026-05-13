package application

import (
	"context"

	"github.com/czanix/boilerplate-api-go/internal/domain"
)

type CreateOrderInput struct {
	CustomerID string             `json:"customerId" binding:"required"`
	Items      []OrderItemInput   `json:"items" binding:"required,min=1"`
}

type OrderItemInput struct {
	ProductID string  `json:"productId" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	UnitPrice float64 `json:"unitPrice" binding:"required,min=0"`
}

type OrderOutput struct {
	PublicID   string  `json:"publicId"`
	CustomerID string  `json:"customerId"`
	Status     string  `json:"status"`
	Total      float64 `json:"total"`
	CreatedAt  string  `json:"createdAt"`
}

type CreateOrderUseCase struct {
	repo domain.OrderRepository
}

func NewCreateOrderUseCase(repo domain.OrderRepository) *CreateOrderUseCase {
	return &CreateOrderUseCase{repo: repo}
}

func (uc *CreateOrderUseCase) Execute(ctx context.Context, input CreateOrderInput) (*OrderOutput, error) {
	items := make([]domain.OrderItem, len(input.Items))
	for i, item := range input.Items {
		items[i] = domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}
	}

	order, err := domain.NewOrder(input.CustomerID, items)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, order); err != nil {
		return nil, err
	}

	return &OrderOutput{
		PublicID:   order.PublicID,
		CustomerID: order.CustomerID,
		Status:     string(order.Status),
		Total:      order.Total(),
		CreatedAt:  order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
