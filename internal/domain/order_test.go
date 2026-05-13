package domain

import (
	"testing"
)

func TestNewOrder_Success(t *testing.T) {
	items := []OrderItem{{ProductID: "prod-1", Quantity: 2, UnitPrice: 29.90}}
	order, err := NewOrder("customer-123", items)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.Status != StatusPending {
		t.Errorf("expected pending, got %s", order.Status)
	}
	if order.CustomerID != "customer-123" {
		t.Errorf("expected customer-123, got %s", order.CustomerID)
	}
}

func TestNewOrder_EmptyItems(t *testing.T) {
	_, err := NewOrder("customer-123", []OrderItem{})
	if err != ErrEmptyOrder {
		t.Errorf("expected ErrEmptyOrder, got %v", err)
	}
}

func TestNewOrder_InvalidQuantity(t *testing.T) {
	items := []OrderItem{{ProductID: "prod-1", Quantity: -1, UnitPrice: 10}}
	_, err := NewOrder("customer-123", items)
	if err != ErrInvalidQuantity {
		t.Errorf("expected ErrInvalidQuantity, got %v", err)
	}
}

func TestOrder_Total(t *testing.T) {
	items := []OrderItem{
		{ProductID: "prod-1", Quantity: 2, UnitPrice: 10},
		{ProductID: "prod-2", Quantity: 3, UnitPrice: 5},
	}
	order, _ := NewOrder("customer-123", items)
	expected := 35.0
	if order.Total() != expected {
		t.Errorf("expected %.2f, got %.2f", expected, order.Total())
	}
}

func TestOrder_Cancel(t *testing.T) {
	items := []OrderItem{{ProductID: "prod-1", Quantity: 1, UnitPrice: 10}}
	order, _ := NewOrder("customer-123", items)
	if err := order.Cancel(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.Status != StatusCancelled {
		t.Errorf("expected cancelled, got %s", order.Status)
	}
}

func TestOrder_Cancel_Delivered(t *testing.T) {
	items := []OrderItem{{ProductID: "prod-1", Quantity: 1, UnitPrice: 10}}
	order, _ := NewOrder("customer-123", items)
	order.Status = StatusDelivered
	if err := order.Cancel(); err != ErrDeliveredCancel {
		t.Errorf("expected ErrDeliveredCancel, got %v", err)
	}
}
