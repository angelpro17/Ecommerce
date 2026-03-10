package postgres

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// NewConnection crea un pool de conexiones a PostgreSQL.
func NewConnection(dsn string, maxOpen, maxIdle int, maxIdleTime time.Duration) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxIdleTime(maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}
