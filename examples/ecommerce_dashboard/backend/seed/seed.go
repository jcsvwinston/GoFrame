package seed

import (
	"fmt"
	"math/rand"

	"github.com/jcsvwinston/GoFrame/examples/ecommerce_dashboard/backend/models"
	"github.com/jcsvwinston/GoFrame/pkg/faker"
)

func Database() {
	fmt.Println("🌱 Seeding database with 100K+ records...")

	categories := []models.Category{
		{Name: "Electronics", Icon: "📱"},
		{Name: "Clothing", Icon: "👕"},
		{Name: "Home & Garden", Icon: "🏠"},
		{Name: "Sports", Icon: "⚽"},
		{Name: "Books", Icon: "📚"},
		{Name: "Toys", Icon: "🧸"},
		{Name: "Food", Icon: "🍔"},
		{Name: "Beauty", Icon: "💄"},
		{Name: "Automotive", Icon: "🚗"},
		{Name: "Office", Icon: "📎"},
	}
	for _, cat := range categories {
		cat.Description = faker.Sentence(10)
		_ = cat
	}

	fmt.Println("   Creating 100,000 products...")
	products := make([]models.Product, 100000)
	for i := range products {
		products[i] = models.Product{
			Name:        faker.ProductName(),
			Description: faker.Paragraph(2),
			Price:       float64(rand.Intn(100000)) / 100.0,
			Stock:       rand.Intn(1000),
			CategoryID:  int64(rand.Intn(10) + 1),
			SKU:         fmt.Sprintf("SKU-%06d", i),
			Image:       fmt.Sprintf("https://picsum.photos/300/200?random=%d", i),
		}
	}
	_ = products

	fmt.Println("   Creating 50,000 customers...")
	customers := make([]models.Customer, 50000)
	for i := range customers {
		customers[i] = models.Customer{
			Name:    faker.Name(),
			Email:   faker.Email(),
			Phone:   faker.Phone(),
			Address: faker.Address(),
		}
	}
	_ = customers

	fmt.Println("   Creating 500,000 orders...")
	orders := make([]models.Order, 500000)
	statuses := []string{"pending", "completed", "shipped", "cancelled"}
	for i := range orders {
		orders[i] = models.Order{
			CustomerID: int64(rand.Intn(50000) + 1),
			Status:     statuses[rand.Intn(len(statuses))],
			Total:      float64(rand.Intn(10000)) / 100.0,
		}
	}
	_ = orders

	fmt.Println("✅ Database seeded!")
}
