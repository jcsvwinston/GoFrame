package quark

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// User is an example model.
type User struct {
	ID        int64     `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	Active    bool      `db:"active" json:"active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func setupTestDB(t *testing.T) (*Client, func()) {
	// Open SQLite in-memory database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// Create table
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL,
			name TEXT,
			active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	// Create quark client
	client, err := New(db, WithDialect(SQLite()))
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	cleanup := func() {
		client.Close()
	}

	return client, cleanup
}

func TestCreate(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user := User{Email: "alice@example.com", Name: "Alice", Active: true}
	err := For[User](ctx, client).Create(&user)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("expected ID to be set after create")
	}

	fmt.Printf("✓ Created user with ID: %d\n", user.ID)
}

func TestFind(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create user
	user := User{Email: "bob@example.com", Name: "Bob", Active: true}
	err := For[User](ctx, client).Create(&user)
	if err != nil {
		t.Fatal(err)
	}

	// Find by ID
	found, err := For[User](ctx, client).Find(user.ID)
	if err != nil {
		t.Fatalf("find user: %v", err)
	}

	if found.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, found.Email)
	}

	fmt.Printf("✓ Found user by ID: %d (%s)\n", found.ID, found.Email)
}

func TestList(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create users
	users := []User{
		{Email: "alice@example.com", Name: "Alice", Active: true},
		{Email: "bob@example.com", Name: "Bob", Active: true},
		{Email: "charlie@example.com", Name: "Charlie", Active: false},
	}

	for i := range users {
		err := For[User](ctx, client).Create(&users[i])
		if err != nil {
			t.Fatal(err)
		}
	}

	// List all
	all, err := For[User](ctx, client).Limit(100).List()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 users, got %d", len(all))
	}

	// List active only
	active, err := For[User](ctx, client).Where("active", "=", true).List()
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 2 {
		t.Errorf("expected 2 active users, got %d", len(active))
	}

	fmt.Printf("✓ Listed %d total users, %d active\n", len(all), len(active))
}

func TestUpdate(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create user
	user := User{Email: "dave@example.com", Name: "Dave", Active: true}
	err := For[User](ctx, client).Create(&user)
	if err != nil {
		t.Fatal(err)
	}

	// Update (only non-zero fields)
	user.Name = "David" // Only changing name
	rows, err := For[User](ctx, client).Update(&user)
	if err != nil {
		t.Fatalf("update user: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}

	// Verify update
	found, err := For[User](ctx, client).Find(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if found.Name != "David" {
		t.Errorf("expected name David, got %s", found.Name)
	}
	if found.Email != "dave@example.com" {
		t.Errorf("email should not change, got %s", found.Email)
	}

	fmt.Printf("✓ Updated user: name changed to %s\n", found.Name)
}

func TestUpdateMap(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create user
	user := User{Email: "eve@example.com", Name: "Eve", Active: true}
	err := For[User](ctx, client).Create(&user)
	if err != nil {
		t.Fatal(err)
	}

	// Bulk update with map
	rows, err := For[User](ctx, client).
		Where("id", "=", user.ID).
		UpdateMap(map[string]any{
			"name":   "Evelyn",
			"active": false,
		})
	if err != nil {
		t.Fatalf("update map: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}

	// Verify update
	found, err := For[User](ctx, client).Find(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if found.Name != "Evelyn" {
		t.Errorf("expected name Evelyn, got %s", found.Name)
	}
	if found.Active != false {
		t.Errorf("expected active false, got %v", found.Active)
	}

	fmt.Printf("✓ Bulk updated user via map\n")
}

func TestDelete(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create user
	user := User{Email: "frank@example.com", Name: "Frank", Active: true}
	err := For[User](ctx, client).Create(&user)
	if err != nil {
		t.Fatal(err)
	}

	// Hard delete (no deleted_at field = hard delete)
	rows, err := For[User](ctx, client).HardDelete(&user)
	if err != nil {
		t.Fatalf("delete user: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row deleted, got %d", rows)
	}

	// Verify deletion
	_, err = For[User](ctx, client).Find(user.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}

	fmt.Printf("✓ Deleted user (hard delete)\n")
}

func TestDeleteBy(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create users
	for i := 0; i < 3; i++ {
		user := User{Email: fmt.Sprintf("user%d@test.com", i), Active: i < 2}
		err := For[User](ctx, client).Create(&user)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Delete inactive users
	rows, err := For[User](ctx, client).
		Where("active", "=", false).
		DeleteBy()
	if err != nil {
		t.Fatalf("delete by: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row deleted, got %d", rows)
	}

	// Verify
	remaining, err := For[User](ctx, client).List()
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 2 {
		t.Errorf("expected 2 remaining users, got %d", len(remaining))
	}

	fmt.Printf("✓ Deleted %d inactive users\n", rows)
}

func TestIter(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create 1000 users
	for i := 0; i < 1000; i++ {
		user := User{Email: fmt.Sprintf("user%d@test.com", i), Name: fmt.Sprintf("User %d", i)}
		if err := For[User](ctx, client).Create(&user); err != nil {
			t.Fatal(err)
		}
	}

	// Use Iter to count without loading all into memory
	count := 0
	err := For[User](ctx, client).Iter(func(user User) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("iter failed: %v", err)
	}
	if count != 1000 {
		t.Errorf("expected 1000 users, got %d", count)
	}

	fmt.Printf("✓ Iter() processed %d users without OOM\n", count)
}

func TestCursor(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create users
	for i := 0; i < 100; i++ {
		user := User{Email: fmt.Sprintf("cursor%d@test.com", i), Name: fmt.Sprintf("Cursor %d", i)}
		if err := For[User](ctx, client).Create(&user); err != nil {
			t.Fatal(err)
		}
	}

	// Use Cursor for manual iteration
	cursor, err := For[User](ctx, client).Where("email", "LIKE", "cursor%").Cursor()
	if err != nil {
		t.Fatal(err)
	}
	defer cursor.Close()

	count := 0
	for cursor.Next() {
		var user User
		if err := cursor.Scan(&user); err != nil {
			t.Fatal(err)
		}
		count++
	}

	if err := cursor.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 100 {
		t.Errorf("expected 100 users, got %d", count)
	}

	fmt.Printf("✓ Cursor() processed %d users\n", count)
}

func TestPaginate(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create 250 users
	for i := 0; i < 250; i++ {
		user := User{Email: fmt.Sprintf("page%d@test.com", i), Name: fmt.Sprintf("Page %d", i)}
		if err := For[User](ctx, client).Create(&user); err != nil {
			t.Fatal(err)
		}
	}

	// Page 0, 100 per page
	page0, err := For[User](ctx, client).OrderBy("id", "ASC").Paginate(100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(page0.Items) != 100 {
		t.Errorf("expected 100 items, got %d", len(page0.Items))
	}
	if page0.Total != 250 {
		t.Errorf("expected total 250, got %d", page0.Total)
	}
	if page0.TotalPages != 3 {
		t.Errorf("expected 3 pages, got %d", page0.TotalPages)
	}

	// Page 2 (last page, only 50 items)
	page2, err := For[User](ctx, client).OrderBy("id", "ASC").Paginate(100, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2.Items) != 50 {
		t.Errorf("expected 50 items on last page, got %d", len(page2.Items))
	}

	fmt.Printf("✓ Paginate() works: %d items, %d pages\n", page0.Total, page0.TotalPages)
}

func TestCount(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create users with different active status
	for i := 0; i < 100; i++ {
		user := User{
			Email:  fmt.Sprintf("count%d@test.com", i),
			Name:   fmt.Sprintf("Count %d", i),
			Active: i%2 == 0, // 50 active, 50 inactive
		}
		if err := For[User](ctx, client).Create(&user); err != nil {
			t.Fatal(err)
		}
	}

	// Count all
	total, err := For[User](ctx, client).Count()
	if err != nil {
		t.Fatal(err)
	}
	if total != 100 {
		t.Errorf("expected 100 total, got %d", total)
	}

	// Count active only
	active, err := For[User](ctx, client).Where("active", "=", true).Count()
	if err != nil {
		t.Fatal(err)
	}
	if active != 50 {
		t.Errorf("expected 50 active, got %d", active)
	}

	fmt.Printf("✓ Count() works: %d total, %d active\n", total, active)
}

func TestQueryBuilder(t *testing.T) {
	client, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create users with different created_at
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	users := []User{
		{Email: "old1@test.com", Name: "Old 1", CreatedAt: baseTime},
		{Email: "old2@test.com", Name: "Old 2", CreatedAt: baseTime.Add(24 * time.Hour)},
		{Email: "new@test.com", Name: "New", CreatedAt: baseTime.Add(7 * 24 * time.Hour)},
	}

	for i := range users {
		err := For[User](ctx, client).Create(&users[i])
		if err != nil {
			t.Fatal(err)
		}
	}

	// Test OrderBy and Limit
	results, err := For[User](ctx, client).
		OrderBy("created_at", "DESC").
		Limit(2).
		List()
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	// Should be newest first
	if results[0].Name != "New" {
		t.Errorf("expected first to be 'New', got %s", results[0].Name)
	}

	fmt.Printf("✓ Query builder with OrderBy and Limit works\n")
}
