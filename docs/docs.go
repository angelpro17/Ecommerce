// Package docs - Swagger para Ecommerce API.
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v1/products": {
            "get": {
                "description": "Obtiene todos los productos con paginacion",
                "produces": ["application/json"],
                "tags": ["products"],
                "summary": "Listar productos",
                "parameters": [
                    {"type": "integer", "default": 1, "description": "Pagina", "name": "page", "in": "query"},
                    {"type": "integer", "default": 10, "description": "Items por pagina", "name": "page_size", "in": "query"}
                ],
                "responses": {
                    "200": {"description": "OK"},
                    "500": {"description": "Internal Server Error"}
                }
            },
            "post": {
                "description": "Crea un nuevo producto",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["products"],
                "summary": "Crear producto",
                "parameters": [
                    {
                        "description": "Datos del producto",
                        "name": "product",
                        "in": "body",
                        "required": true,
                        "schema": {"$ref": "#/definitions/CreateProductRequest"}
                    }
                ],
                "responses": {
                    "201": {"description": "Created"},
                    "400": {"description": "Bad Request"},
                    "500": {"description": "Internal Server Error"}
                }
            }
        },
        "/api/v1/products/{id}": {
            "get": {
                "description": "Obtiene un producto por su ID",
                "produces": ["application/json"],
                "tags": ["products"],
                "summary": "Obtener producto",
                "parameters": [
                    {"type": "integer", "description": "Product ID", "name": "id", "in": "path", "required": true}
                ],
                "responses": {
                    "200": {"description": "OK"},
                    "400": {"description": "Bad Request"},
                    "404": {"description": "Not Found"}
                }
            },
            "put": {
                "description": "Actualiza un producto existente",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["products"],
                "summary": "Actualizar producto",
                "parameters": [
                    {"type": "integer", "description": "Product ID", "name": "id", "in": "path", "required": true},
                    {
                        "description": "Campos a actualizar",
                        "name": "product",
                        "in": "body",
                        "required": true,
                        "schema": {"$ref": "#/definitions/UpdateProductRequest"}
                    }
                ],
                "responses": {
                    "200": {"description": "OK"},
                    "400": {"description": "Bad Request"},
                    "404": {"description": "Not Found"},
                    "500": {"description": "Internal Server Error"}
                }
            },
            "delete": {
                "description": "Elimina un producto por su ID",
                "tags": ["products"],
                "summary": "Eliminar producto",
                "parameters": [
                    {"type": "integer", "description": "Product ID", "name": "id", "in": "path", "required": true}
                ],
                "responses": {
                    "204": {"description": "No Content"},
                    "400": {"description": "Bad Request"},
                    "404": {"description": "Not Found"},
                    "500": {"description": "Internal Server Error"}
                }
            }
        },
        "/api/v1/cart": {
            "get": {
                "description": "Obtiene el resumen del carrito del usuario",
                "produces": ["application/json"],
                "tags": ["cart"],
                "summary": "Ver carrito",
                "parameters": [
                    {"type": "string", "description": "ID del usuario", "name": "X-User-ID", "in": "header"}
                ],
                "responses": {
                    "200": {"description": "OK"},
                    "500": {"description": "Internal Server Error"}
                }
            },
            "delete": {
                "description": "Elimina todos los items del carrito",
                "tags": ["cart"],
                "summary": "Vaciar carrito",
                "parameters": [
                    {"type": "string", "description": "ID del usuario", "name": "X-User-ID", "in": "header"}
                ],
                "responses": {
                    "204": {"description": "No Content"},
                    "404": {"description": "Not Found"},
                    "500": {"description": "Internal Server Error"}
                }
            }
        },
        "/api/v1/cart/items": {
            "post": {
                "description": "Agrega un producto al carrito",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["cart"],
                "summary": "Agregar item al carrito",
                "parameters": [
                    {"type": "string", "description": "ID del usuario", "name": "X-User-ID", "in": "header"},
                    {
                        "description": "Producto y cantidad",
                        "name": "item",
                        "in": "body",
                        "required": true,
                        "schema": {"$ref": "#/definitions/AddItemRequest"}
                    }
                ],
                "responses": {
                    "200": {"description": "OK"},
                    "400": {"description": "Bad Request"},
                    "404": {"description": "Not Found"},
                    "409": {"description": "Conflict - Stock insuficiente"}
                }
            }
        },
        "/api/v1/cart/items/{productId}": {
            "put": {
                "description": "Actualiza la cantidad de un item en el carrito",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["cart"],
                "summary": "Actualizar cantidad de item",
                "parameters": [
                    {"type": "string", "description": "ID del usuario", "name": "X-User-ID", "in": "header"},
                    {"type": "integer", "description": "Product ID", "name": "productId", "in": "path", "required": true},
                    {
                        "description": "Nueva cantidad",
                        "name": "item",
                        "in": "body",
                        "required": true,
                        "schema": {"$ref": "#/definitions/UpdateItemRequest"}
                    }
                ],
                "responses": {
                    "200": {"description": "OK"},
                    "400": {"description": "Bad Request"},
                    "404": {"description": "Not Found"},
                    "409": {"description": "Conflict - Stock insuficiente"}
                }
            },
            "delete": {
                "description": "Elimina un item del carrito",
                "tags": ["cart"],
                "summary": "Eliminar item del carrito",
                "parameters": [
                    {"type": "string", "description": "ID del usuario", "name": "X-User-ID", "in": "header"},
                    {"type": "integer", "description": "Product ID", "name": "productId", "in": "path", "required": true}
                ],
                "responses": {
                    "200": {"description": "OK"},
                    "400": {"description": "Bad Request"},
                    "404": {"description": "Not Found"},
                    "500": {"description": "Internal Server Error"}
                }
            }
        },
        "/api/v1/cart/checkout": {
            "post": {
                "description": "Procesa el checkout del carrito",
                "produces": ["application/json"],
                "tags": ["cart"],
                "summary": "Checkout",
                "parameters": [
                    {"type": "string", "description": "ID del usuario", "name": "X-User-ID", "in": "header"}
                ],
                "responses": {
                    "200": {"description": "OK"},
                    "500": {"description": "Internal Server Error"}
                }
            }
        }
    },
    "definitions": {
        "CreateProductRequest": {
            "type": "object",
            "required": ["name", "price"],
            "properties": {
                "name": {"type": "string", "example": "Laptop"},
                "description": {"type": "string", "example": "Laptop gaming"},
                "price": {"type": "number", "example": 999.99},
                "stock": {"type": "integer", "example": 10},
                "image_url": {"type": "string", "example": "https://img.com/laptop.jpg"}
            }
        },
        "UpdateProductRequest": {
            "type": "object",
            "properties": {
                "name": {"type": "string"},
                "description": {"type": "string"},
                "price": {"type": "number"},
                "stock": {"type": "integer"},
                "image_url": {"type": "string"}
            }
        },
        "AddItemRequest": {
            "type": "object",
            "required": ["product_id", "quantity"],
            "properties": {
                "product_id": {"type": "integer", "example": 1},
                "quantity": {"type": "integer", "example": 2}
            }
        },
        "UpdateItemRequest": {
            "type": "object",
            "required": ["quantity"],
            "properties": {
                "quantity": {"type": "integer", "example": 3}
            }
        }
    }
}`

var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Ecommerce API",
	Description:      "API REST de ecommerce con productos y carrito de compras",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
