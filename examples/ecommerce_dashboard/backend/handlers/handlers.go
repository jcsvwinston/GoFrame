package handlers

import (
	"math/rand"

	"github.com/jcsvwinston/GoFrame/examples/ecommerce_dashboard/backend/models"
	"github.com/jcsvwinston/GoFrame/pkg/fluent"
)

func GetStats(c *fluent.Context) error {
	stats := models.StatsResponse{
		TotalProducts:  100000,
		TotalOrders:    500000,
		TotalCustomers: 50000,
		Revenue:        2500000.50,
		OrdersToday:    rand.Int63n(500),
		RevenueToday:   float64(rand.Int63n(50000)) / 100.0,
	}
	return c.JSON(200, stats)
}

func ListProducts(c *fluent.Context) error {
	return c.JSON(200, map[string]interface{}{
		"products": []models.Product{
			{Name: "Sample Product", Price: 99.99, Stock: 100},
		},
		"total": 100000,
	})
}

func CreateProduct(c *fluent.Context) error {
	var product models.Product
	if err := c.BindJSON(&product); err != nil {
		return err
	}
	return c.JSON(201, product)
}

func GetProduct(c *fluent.Context) error {
	id := c.Param("id")
	return c.JSON(200, models.Product{
		Name:  "Product " + id,
		Price: 99.99,
	})
}

func ListOrders(c *fluent.Context) error {
	return c.JSON(200, map[string]interface{}{
		"orders": []models.Order{},
		"total":  500000,
	})
}

func CreateOrder(c *fluent.Context) error {
	var order models.Order
	if err := c.BindJSON(&order); err != nil {
		return err
	}
	return c.JSON(201, order)
}

func ListCustomers(c *fluent.Context) error {
	return c.JSON(200, map[string]interface{}{
		"customers": []models.Customer{},
		"total":     50000,
	})
}

func GetCustomer(c *fluent.Context) error {
	id := c.Param("id")
	return c.JSON(200, models.Customer{
		Name:  "Customer " + id,
		Email: "customer@example.com",
	})
}

func ListCategories(c *fluent.Context) error {
	return c.JSON(200, []models.Category{
		{Name: "Electronics", Icon: "📱"},
		{Name: "Clothing", Icon: "👕"},
		{Name: "Home", Icon: "🏠"},
	})
}
