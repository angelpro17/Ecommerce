package postgres

import (
	"context"
	"database/sql"
)

// RunMigrations ejecuta las migraciones del esquema de la base de datos.
func RunMigrations(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		price DECIMAL(10,2) NOT NULL,
		stock INTEGER NOT NULL DEFAULT 0,
		image_url VARCHAR(500),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS carts (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL,
		status VARCHAR(50) DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS cart_items (
		id SERIAL PRIMARY KEY,
		cart_id INTEGER NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
		product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
		quantity INTEGER NOT NULL DEFAULT 1,
		price DECIMAL(10,2) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(cart_id, product_id)
	);
	CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id ON cart_items(cart_id);
	CREATE INDEX IF NOT EXISTS idx_carts_user_id ON carts(user_id);
	`)
	return err
}
