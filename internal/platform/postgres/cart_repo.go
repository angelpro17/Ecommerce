package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/angelpro17/Ecommerce.git/internal/cart/domain"
)

// CartPostgresRepo implementa domain.Repository con PostgreSQL.
type CartPostgresRepo struct {
	db *sql.DB
}

func NewCartPostgresRepo(db *sql.DB) *CartPostgresRepo {
	return &CartPostgresRepo{db: db}
}

func (r *CartPostgresRepo) Create(ctx context.Context, userID string) (*domain.Cart, error) {
	now := time.Now()
	cart := &domain.Cart{UserID: userID, Status: "active", CreatedAt: now, UpdatedAt: now}
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO carts (user_id,status,created_at,updated_at) VALUES ($1,'active',$2,$3) RETURNING id`,
		userID, now, now).Scan(&cart.ID)
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (r *CartPostgresRepo) GetByID(ctx context.Context, id int64) (*domain.Cart, error) {
	c := &domain.Cart{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id,user_id,status,created_at,updated_at FROM carts WHERE id=$1`, id,
	).Scan(&c.ID, &c.UserID, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrCartNotFound
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CartPostgresRepo) GetActiveByUserID(ctx context.Context, userID string) (*domain.Cart, error) {
	c := &domain.Cart{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id,user_id,status,created_at,updated_at FROM carts
		 WHERE user_id=$1 AND status='active' ORDER BY created_at DESC LIMIT 1`, userID,
	).Scan(&c.ID, &c.UserID, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrCartNotFound
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CartPostgresRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE carts SET status=$1, updated_at=$2 WHERE id=$3`, status, time.Now(), id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrCartNotFound
	}
	return nil
}

func (r *CartPostgresRepo) AddItem(ctx context.Context, item *domain.CartItem) error {
	now := time.Now()
	item.CreatedAt, item.UpdatedAt = now, now
	return r.db.QueryRowContext(ctx,
		`INSERT INTO cart_items (cart_id,product_id,quantity,price,created_at,updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (cart_id,product_id)
		 DO UPDATE SET quantity=cart_items.quantity+EXCLUDED.quantity, updated_at=EXCLUDED.updated_at
		 RETURNING id`,
		item.CartID, item.ProductID, item.Quantity, item.Price, now, now).Scan(&item.ID)
}

func (r *CartPostgresRepo) UpdateItemQuantity(ctx context.Context, cartID, productID int64, qty int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE cart_items SET quantity=$1, updated_at=$2 WHERE cart_id=$3 AND product_id=$4`,
		qty, time.Now(), cartID, productID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrItemNotFound
	}
	return nil
}

func (r *CartPostgresRepo) RemoveItem(ctx context.Context, cartID, productID int64) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM cart_items WHERE cart_id=$1 AND product_id=$2`, cartID, productID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrItemNotFound
	}
	return nil
}

func (r *CartPostgresRepo) GetItems(ctx context.Context, cartID int64) ([]domain.CartItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT ci.id,ci.cart_id,ci.product_id,p.name,ci.quantity,ci.price,ci.created_at,ci.updated_at
		 FROM cart_items ci JOIN products p ON ci.product_id=p.id WHERE ci.cart_id=$1`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.CartItem
	for rows.Next() {
		var it domain.CartItem
		if err := rows.Scan(&it.ID, &it.CartID, &it.ProductID, &it.Name, &it.Quantity, &it.Price, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, err
		}
		it.Subtotal = it.Price * float64(it.Quantity)
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *CartPostgresRepo) ClearCart(ctx context.Context, cartID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM cart_items WHERE cart_id=$1`, cartID)
	return err
}
