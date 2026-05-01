package main

import (
	"fmt"

	"github.com/jcsvwinston/GoFrame/examples/ecommerce_dashboard/backend/handlers"
	"github.com/jcsvwinston/GoFrame/examples/ecommerce_dashboard/backend/models"
	"github.com/jcsvwinston/GoFrame/examples/ecommerce_dashboard/backend/seed"
	"github.com/jcsvwinston/GoFrame/pkg/fluent"
)

func main() {
	app := fluent.New().
		Port(8080).
		SQLite("ecommerce.db").
		Model(&models.Product{}).
		Model(&models.Category{}).
		Model(&models.Order{}).
		Model(&models.OrderItem{}).
		Model(&models.Customer{}).
		AutoMigrate()

	seed.Database()

	api := app.Group("/api")
	api.Get("/stats", handlers.GetStats)
	api.Get("/products", handlers.ListProducts)
	api.Post("/products", handlers.CreateProduct)
	api.Get("/products/:id", handlers.GetProduct)
	api.Get("/orders", handlers.ListOrders)
	api.Post("/orders", handlers.CreateOrder)
	api.Get("/customers", handlers.ListCustomers)
	api.Get("/customers/:id", handlers.GetCustomer)
	api.Get("/categories", handlers.ListCategories)

	app.SPA("../frontend/dist", fluent.SPAConfig{
		IndexFile: "index.html",
		APIPrefix: "/api",
	})

	fmt.Println("🚀 E-Commerce Dashboard")
	fmt.Println("📊 API: http://localhost:8080/api")
	fmt.Println("🌐 App: http://localhost:8080")

	if err := app.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
