package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	cartApp "github.com/angelpro17/Ecommerce.git/internal/cart/application"
	cartHTTP "github.com/angelpro17/Ecommerce.git/internal/cart/interfaces/http"
	"github.com/angelpro17/Ecommerce.git/internal/platform/postgres"
	productApp "github.com/angelpro17/Ecommerce.git/internal/product/application"
	productHTTP "github.com/angelpro17/Ecommerce.git/internal/product/interfaces/http"

	_ "github.com/angelpro17/Ecommerce.git/docs"
)

func main() {
	cfg := config{
		addr: getEnv("SERVER_ADDR", ":8080"),
		db: dbConfig{
			dsn:          getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable"),
			maxOpenConns: 25,
			maxIdleConns: 25,
			maxIdleTime:  15 * time.Minute,
		},
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	// Infraestructura
	db, err := postgres.NewConnection(cfg.db.dsn, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}
	defer db.Close()

	if err := postgres.RunMigrations(context.Background(), db); err != nil {
		log.Fatalf("Error ejecutando migraciones: %v", err)
	}
	log.Println("Base de datos conectada y migrada")

	// Repositorios
	productRepo := postgres.NewProductPostgresRepo(db)
	cartRepo := postgres.NewCartPostgresRepo(db)
	productReader := postgres.NewProductReaderAdapter(productRepo)

	// Servicios
	productSvc := productApp.NewProductService(productRepo)
	cartSvc := cartApp.NewCartService(cartRepo, productReader)

	// Servidor HTTP
	app := application{
		config:         cfg,
		db:             db,
		productHandler: productHTTP.NewHandler(productSvc),
		cartHandler:    cartHTTP.NewHandler(cartSvc),
	}
	if err := app.run(app.mount()); err != nil {
		slog.Error("Error al iniciar servidor", "error", err)
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
