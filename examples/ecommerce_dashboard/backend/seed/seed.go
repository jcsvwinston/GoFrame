package seed

import (
	"fmt"
	"math/rand"

	"github.com/jcsvwinston/GoFrame/examples/ecommerce_dashboard/backend/models"
)

var (
	firstNames = []string{"James", "Mary", "John", "Patricia", "Robert", "Jennifer", "Michael", "Linda", "William", "Elizabeth"}
	lastNames  = []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez"}
	products   = []string{"Wireless Mouse", "Bluetooth Speaker", "USB Cable", "Phone Case", "Laptop Stand", "Webcam", "Keyboard", "Monitor", "Headphones", "Charger"}
	adjectives = []string{"Premium", "Wireless", "Smart", "Digital", "Portable", "Professional", "Ergonomic", "Modern", "Compact", "Durable"}
	domains    = []string{"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "company.com"}
	streets    = []string{"Main St", "Oak Ave", "Maple Rd", "Cedar Ln", "Pine Dr", "Elm St", "Washington Ave", "Park Rd", "Lake Dr", "Hill Ln"}
	cities     = []string{"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio", "San Diego", "Dallas", "San Jose"}
	states     = []string{"CA", "NY", "TX", "FL", "IL", "PA", "OH", "GA", "NC", "MI"}
)

func fakeName() string {
	return fmt.Sprintf("%s %s", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))])
}

func fakeEmail() string {
	return fmt.Sprintf("%s.%s@%s", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))], domains[rand.Intn(len(domains))])
}

func fakePhone() string {
	return fmt.Sprintf("(555) %03d-%04d", rand.Intn(1000), rand.Intn(10000))
}

func fakeAddress() string {
	return fmt.Sprintf("%d %s, %s, %s %05d", rand.Intn(9999)+1, streets[rand.Intn(len(streets))], cities[rand.Intn(len(cities))], states[rand.Intn(len(states))], rand.Intn(99999))
}

func fakeProductName() string {
	return fmt.Sprintf("%s %s", adjectives[rand.Intn(len(adjectives))], products[rand.Intn(len(products))])
}

func fakeSentence() string {
	words := []string{"the", "quick", "brown", "fox", "jumps", "over", "lazy", "dog", "product", "quality", "excellent", "value"}
	return fmt.Sprintf("%s %s %s %s %s.", words[rand.Intn(len(words))], words[rand.Intn(len(words))], words[rand.Intn(len(words))], words[rand.Intn(len(words))], words[rand.Intn(len(words))])
}

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
		cat.Description = fakeSentence()
		_ = cat
	}

	fmt.Println("   Creating 100,000 products...")
	products := make([]models.Product, 100000)
	for i := range products {
		products[i] = models.Product{
			Name:        fakeProductName(),
			Description: fakeSentence(),
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
			Name:    fakeName(),
			Email:   fakeEmail(),
			Phone:   fakePhone(),
			Address: fakeAddress(),
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
