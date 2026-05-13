package postgres

import (
	"context"

	"github.com/czanix/boilerplate-api-go/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgOrderRepository struct {
	pool *pgxpool.Pool
}

func NewPgOrderRepository(pool *pgxpool.Pool) *PgOrderRepository {
	return &PgOrderRepository{pool: pool}
}

func (r *PgOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var orderID int64
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (public_id, customer_id, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		order.PublicID, order.CustomerID, order.Status, order.CreatedAt, order.UpdatedAt,
	).Scan(&orderID)
	if err != nil {
		return err
	}

	for _, item := range order.Items {
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES ($1, $2, $3, $4)`,
			orderID, item.ProductID, item.Quantity, item.UnitPrice,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PgOrderRepository) FindByPublicID(ctx context.Context, publicID string) (*domain.Order, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, public_id, customer_id, status, created_at, updated_at
		 FROM orders WHERE public_id = $1 AND deleted_at IS NULL`, publicID)

	var o domain.Order
	err := row.Scan(&o.ID, &o.PublicID, &o.CustomerID, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT product_id, quantity, unit_price FROM order_items WHERE order_id = $1`, o.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.UnitPrice); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, item)
	}

	return &o, nil
}

func (r *PgOrderRepository) FindByCustomerID(ctx context.Context, customerID string) ([]*domain.Order, error) {
	return nil, nil // implement as needed
}

func (r *PgOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE orders SET status = $1, updated_at = NOW() WHERE public_id = $2`,
		order.Status, order.PublicID)
	return err
}

func (r *PgOrderRepository) SoftDelete(ctx context.Context, publicID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE orders SET deleted_at = NOW(), updated_at = NOW() WHERE public_id = $1`, publicID)
	return err
}
