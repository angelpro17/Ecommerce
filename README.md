# Ecommerce API

API REST de ecommerce construida con **Go**, **Chi Router** y **PostgreSQL**, siguiendo **Clean Architecture**.

## Tecnologias

- **Go 1.21+** - Lenguaje principal
- **Chi v5** - Router HTTP
- **PostgreSQL 14+** - Base de datos
- **Swagger** - Documentacion interactiva

## Arquitectura

```
internal/
├── product/                    # Modulo de productos
│   ├── domain/                 # Entidades, errores y puertos (interfaces)
│   ├── application/            # Logica de negocio (service)
│   └── interfaces/http/        # Handlers HTTP y DTOs
├── cart/                       # Modulo de carrito
│   ├── domain/
│   ├── application/
│   └── interfaces/http/
└── platform/                   # Infraestructura compartida
    ├── postgres/               # Repositorios, migraciones, conexion
    └── http/                   # Helpers de respuesta JSON
```

Cada modulo sigue la regla de dependencia: `domain` no importa nada externo, `application` solo depende de `domain`, y `interfaces` + `platform` implementan los puertos.

## Endpoints

### Products

| Metodo | Ruta | Descripcion |
|--------|------|-------------|
| GET | `/api/v1/products` | Listar (paginado) |
| GET | `/api/v1/products/{id}` | Obtener por ID |
| POST | `/api/v1/products` | Crear |
| PUT | `/api/v1/products/{id}` | Actualizar |
| DELETE | `/api/v1/products/{id}` | Eliminar |

### Cart

| Metodo | Ruta | Descripcion |
|--------|------|-------------|
| GET | `/api/v1/cart` | Ver carrito |
| POST | `/api/v1/cart/items` | Agregar item |
| PUT | `/api/v1/cart/items/{productId}` | Actualizar cantidad |
| DELETE | `/api/v1/cart/items/{productId}` | Eliminar item |
| DELETE | `/api/v1/cart` | Vaciar carrito |
| POST | `/api/v1/cart/checkout` | Checkout |

> El header `X-User-ID` identifica al usuario. Si no se envia, se usa `anonymous`.

## Ejecutar

```bash
# PostgreSQL con Docker
docker run --name postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=ecommerce -p 5432:5432 -d postgres:15

# Iniciar servidor
go run cmd/*.go
```

El servidor inicia en `http://localhost:8080`. Swagger UI en `/swagger/index.html`.

### Variables de entorno

| Variable | Default |
|----------|---------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable` |
| `SERVER_ADDR` | `:8080` |

## Tests

```bash
go test ./internal/... -cover
```

| Paquete | Cobertura |
|---------|-----------|
| `cart/application` | 98.6% |
| `cart/interfaces/http` | 98.6% |
| `product/application` | 97.5% |
| `product/interfaces/http` | 100% |

## Funcionalidades

- CRUD completo de productos con validaciones
- Carrito de compras por usuario con calculo de IVA (16%)
- Control de stock en tiempo real
- Checkout con actualizacion de inventario
- Paginacion en listados
- Migraciones automaticas
- CORS habilitado
- Swagger UI integrado
