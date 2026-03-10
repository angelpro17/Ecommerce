# Ecommerce API - Clean Architecture

API REST de ecommerce con Go, Chi Router y PostgreSQL.

## Estructura del Proyecto (Clean Architecture)

```
├── cmd/
│   ├── main.go          # Punto de entrada, configuración y DI
│   └── api.go           # Configuración del servidor HTTP
├── internal/
│   ├── store/
│   │   ├── postgres.go  # Conexión a PostgreSQL
│   │   └── migrations.go # Migraciones de BD
│   ├── products/
│   │   ├── entity.go    # Entidades del dominio
│   │   ├── repository.go # Interface del repositorio
│   │   ├── postgres.go  # Implementación PostgreSQL
│   │   ├── service.go   # Lógica de negocio
│   │   └── handlers.go  # Handlers HTTP
│   └── cart/
│       ├── entity.go    # Entidades del carrito
│       ├── repository.go # Interface del repositorio
│       ├── postgres.go  # Implementación PostgreSQL
│       ├── service.go   # Lógica de negocio
│       └── handlers.go  # Handlers HTTP
└── go.mod
```

## Iniciar

### 1. Requisitos
- Go 1.21+
- PostgreSQL 14+

### 2. Configurar PostgreSQL

```bash
# Crear base de datos
createdb ecommerce

# O con Docker
docker run --name postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=ecommerce -p 5432:5432 -d postgres:15
```

### 3. Variables de entorno (opcional)

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable"
export SERVER_ADDR=":8080"
```

### 4. Ejecutar

```bash
go run cmd/*.go
```

## API Endpoints

### Products (CRUD)

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/v1/products` | Listar productos (paginado) |
| GET | `/api/v1/products/{id}` | Obtener producto por ID |
| POST | `/api/v1/products` | Crear producto |
| PUT | `/api/v1/products/{id}` | Actualizar producto |
| DELETE | `/api/v1/products/{id}` | Eliminar producto |

### Cart

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/v1/cart` | Ver carrito actual |
| POST | `/api/v1/cart/items` | Agregar item al carrito |
| PUT | `/api/v1/cart/items/{productId}` | Actualizar cantidad |
| DELETE | `/api/v1/cart/items/{productId}` | Eliminar item |
| DELETE | `/api/v1/cart` | Vaciar carrito |
| POST | `/api/v1/cart/checkout` | Finalizar compra |

## Ejemplos de uso

### Crear producto
```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "iPhone 15",
    "description": "Smartphone Apple",
    "price": 999.99,
    "stock": 50,
    "image_url": "https://example.com/iphone.jpg"
  }'
```

### Listar productos
```bash
curl "http://localhost:8080/api/v1/products?page=1&page_size=10"
```

### Agregar al carrito
```bash
curl -X POST http://localhost:8080/api/v1/cart/items \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user123" \
  -d '{
    "product_id": 1,
    "quantity": 2
  }'
```

### Ver carrito con total
```bash
curl http://localhost:8080/api/v1/cart \
  -H "X-User-ID: user123"
```

### Checkout
```bash
curl -X POST http://localhost:8080/api/v1/cart/checkout \
  -H "X-User-ID: user123"
```

## Clean Architecture

```
┌─────────────────────────────────────────────────┐
│                   Handlers                       │  ← Capa de Presentación
│            (HTTP Request/Response)               │
└─────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│                   Service                        │  ← Capa de Negocio
│            (Business Logic)                      │
└─────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│                 Repository                       │  ← Capa de Datos
│         (Interface + PostgreSQL)                 │
└─────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│                  PostgreSQL                      │  ← Base de Datos
└─────────────────────────────────────────────────┘
```

## Características

- CRUD completo de productos
- Carrito de compras por usuario
- Cálculo automático de totales con IVA (16%)
- Control de stock
- Checkout con actualización de inventario
- Paginación en listados
- Migraciones automáticas
- CORS habilitado
- Middlewares (Logger, Recovery, Timeout)
# Ecommerce
