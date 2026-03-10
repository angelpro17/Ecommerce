package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/angelpro17/Ecommerce.git/internal/product/domain"
)

// ProductPostgresRepo implementa domain.Repository con PostgreSQL.
type ProductPostgresRepo struct {
	db *sql.DB
}

func NewProductPostgresRepo(db *sql.DB) *ProductPostgresRepo {
	return &ProductPostgresRepo{db: db}
}

func (r *ProductPostgresRepo) Create(ctx context.Context, p *domain.Product) error {
	now := time.Now()
	p.CreatedAt, p.UpdatedAt = now, now
	return r.db.QueryRowContext(ctx,
		`INSERT INTO products (name,description,price,stock,image_url,created_at,updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		p.Name, p.Description, p.Price, p.Stock, p.ImageURL, now, now,
	).Scan(&p.ID)
}

func (r *ProductPostgresRepo) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	p := &domain.Product{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id,name,description,price,stock,image_url,created_at,updated_at FROM products WHERE id=$1`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProductPostgresRepo) GetAll(ctx context.Context, page, pageSize int) ([]domain.Product, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM products`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id,name,description,price,stock,image_url,created_at,updated_at
		 FROM products ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	return products, total, rows.Err()
}

func (r *ProductPostgresRepo) Update(ctx context.Context, p *domain.Product) error {
	p.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx,
		`UPDATE products SET name=$1,description=$2,price=$3,stock=$4,image_url=$5,updated_at=$6 WHERE id=$7`,
		p.Name, p.Description, p.Price, p.Stock, p.ImageURL, p.UpdatedAt, p.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ProductPostgresRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ProductPostgresRepo) UpdateStock(ctx context.Context, id int64, qty int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE products SET stock=stock+$1, updated_at=$2 WHERE id=$3 AND stock+$1>=0`,
		qty, time.Now(), id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("insufficient stock or product not found")
	}
	return nil
}
