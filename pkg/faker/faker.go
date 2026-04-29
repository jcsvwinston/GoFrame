// Package faker provides utilities for generating fake data for testing and seeding
package faker

import (
	"fmt"
	"math/rand"
	"strings"
)

var (
	firstNames = []string{"James", "Mary", "John", "Patricia", "Robert", "Jennifer", "Michael", "Linda", "William", "Elizabeth", "David", "Barbara", "Richard", "Susan", "Joseph", "Jessica", "Thomas", "Sarah", "Charles", "Karen"}
	lastNames  = []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin"}
	products   = []string{"Wireless Mouse", "Bluetooth Speaker", "USB Cable", "Phone Case", "Laptop Stand", "Webcam", "Keyboard", "Monitor", "Headphones", "Charger", "Adapter", "Mouse Pad", "Desk Lamp", "Notebook", "Pen Set", "Backpack", "Water Bottle", "Sunglasses", "Watch", "Wallet"}
	adjectives = []string{"Premium", "Wireless", "Smart", "Digital", "Portable", "Professional", "Ergonomic", "Modern", "Compact", "Durable", "Stylish", "Lightweight", "Heavy-duty", "Sleek", "Advanced", "Essential", "Multi-purpose", "High-quality", "Budget", "Luxury"}
	domains    = []string{"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "company.com", "business.com"}
	streets    = []string{"Main St", "Oak Ave", "Maple Rd", "Cedar Ln", "Pine Dr", "Elm St", "Washington Ave", "Park Rd", "Lake Dr", "Hill Ln"}
	cities     = []string{"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio", "San Diego", "Dallas", "San Jose"}
	states     = []string{"CA", "NY", "TX", "FL", "IL", "PA", "OH", "GA", "NC", "MI"}
)

// Name generates a random full name
func Name() string {
	return fmt.Sprintf("%s %s", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))])
}

// Email generates a random email address
func Email() string {
	return fmt.Sprintf("%s.%s@%s", strings.ToLower(firstNames[rand.Intn(len(firstNames))]), strings.ToLower(lastNames[rand.Intn(len(lastNames))]), domains[rand.Intn(len(domains))])
}

// Phone generates a random phone number
func Phone() string {
	return fmt.Sprintf("(555) %03d-%04d", rand.Intn(1000), rand.Intn(10000))
}

// Address generates a random address
func Address() string {
	return fmt.Sprintf("%d %s, %s, %s %05d", rand.Intn(9999)+1, streets[rand.Intn(len(streets))], cities[rand.Intn(len(cities))], states[rand.Intn(len(states))], rand.Intn(99999))
}

// ProductName generates a random product name
func ProductName() string {
	return fmt.Sprintf("%s %s", adjectives[rand.Intn(len(adjectives))], products[rand.Intn(len(products))])
}

// Sentence generates a sentence with n words
func Sentence(words int) string {
	if words <= 0 {
		words = 5
	}
	wordList := []string{"the", "quick", "brown", "fox", "jumps", "over", "lazy", "dog", "product", "quality", "excellent", "value", "money", "recommend", "highly", "amazing", "great", "awesome", "fantastic", "wonderful"}
	var result []string
	for i := 0; i < words; i++ {
		result = append(result, wordList[rand.Intn(len(wordList))])
	}
	return strings.Join(result, " ") + "."
}

// Paragraph generates a paragraph with n sentences
func Paragraph(sentences int) string {
	if sentences <= 0 {
		sentences = 3
	}
	var result []string
	for i := 0; i < sentences; i++ {
		result = append(result, Sentence(rand.Intn(10)+5))
	}
	return strings.Join(result, " ")
}

// Price generates a random price between min and max
func Price(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// Int generates a random int between min and max
func Int(min, max int) int {
	return min + rand.Intn(max-min)
}

// UUID generates a random UUID-like string
func UUID() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x", rand.Int63(), rand.Int31(), rand.Int31(), rand.Int31(), rand.Int63())
}
