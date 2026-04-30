package models

import "github.com/jcsvwinston/GoFrame/pkg/model"

type Product struct {
	model.BaseModel
	Name        string  `json:"name" db:"name" validate:"required"`
	Description string  `json:"description" db:"description"`
	Price       float64 `json:"price" db:"price" validate:"min=0"`
	Stock       int     `json:"stock" db:"stock" validate:"min=0"`
	CategoryID  int64   `json:"category_id" db:"category_id"`
	Image       string  `json:"image" db:"image"`
	SKU         string  `json:"sku" db:"sku" validate:"required"`
}

type Category struct {
	model.BaseModel
	Name        string `json:"name" db:"name" validate:"required"`
	Description string `json:"description" db:"description"`
	Icon        string `json:"icon" db:"icon"`
}

type Order struct {
	model.BaseModel
	CustomerID int64       `json:"customer_id" db:"customer_id"`
	Status     string      `json:"status" db:"status" validate:"required"`
	Total      float64     `json:"total" db:"total"`
	Items      []OrderItem `json:"items" db:"-"`
}

type OrderItem struct {
	model.BaseModel
	OrderID   int64   `json:"order_id" db:"order_id"`
	ProductID int64   `json:"product_id" db:"product_id"`
	Quantity  int     `json:"quantity" db:"quantity"`
	Price     float64 `json:"price" db:"price"`
}

type Customer struct {
	model.BaseModel
	Name    string `json:"name" db:"name" validate:"required"`
	Email   string `json:"email" db:"email" validate:"required,email"`
	Phone   string `json:"phone" db:"phone"`
	Address string `json:"address" db:"address"`
}

type StatsResponse struct {
	TotalProducts  int64   `json:"total_products"`
	TotalOrders    int64   `json:"total_orders"`
	TotalCustomers int64   `json:"total_customers"`
	Revenue        float64 `json:"revenue"`
	OrdersToday    int64   `json:"orders_today"`
	RevenueToday   float64 `json:"revenue_today"`
}
