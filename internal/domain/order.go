package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusCancelled OrderStatus = "cancelled"
	StatusDelivered OrderStatus = "delivered"
)

var (
	ErrEmptyOrder      = errors.New("pedido deve ter pelo menos um item")
	ErrInvalidQuantity = errors.New("quantidade deve ser positiva")
	ErrAlreadyCancelled = errors.New("pedido já cancelado")
	ErrDeliveredCancel = errors.New("não é possível cancelar pedido entregue")
)

type OrderItem struct {
	ProductID string
	Quantity  int
	UnitPrice float64
}

type Order struct {
	ID         int64
	PublicID   string
	CustomerID string
	Items      []OrderItem
	Status     OrderStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

func NewOrder(customerID string, items []OrderItem) (*Order, error) {
	if len(items) == 0 {
		return nil, ErrEmptyOrder
	}
	for _, item := range items {
		if item.Quantity <= 0 {
			return nil, ErrInvalidQuantity
		}
	}

	now := time.Now()
	return &Order{
		PublicID:   uuid.New().String(),
		CustomerID: customerID,
		Items:      items,
		Status:     StatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (o *Order) Total() float64 {
	var total float64
	for _, item := range o.Items {
		total += float64(item.Quantity) * item.UnitPrice
	}
	return total
}

func (o *Order) Cancel() error {
	if o.Status == StatusDelivered {
		return ErrDeliveredCancel
	}
	if o.Status == StatusCancelled {
		return ErrAlreadyCancelled
	}
	o.Status = StatusCancelled
	o.UpdatedAt = time.Now()
	return nil
}
