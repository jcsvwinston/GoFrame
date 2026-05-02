package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/sijms/go-ora/v2"
)

func main() {
	dsn := os.Getenv("QUARK_TEST_ORACLE_DSN")
	if dsn == "" {
		log.Fatal("QUARK_TEST_ORACLE_DSN not set")
	}

	db, err := sql.Open("oracle", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Drop and create table
	_, _ = db.ExecContext(ctx, "DROP TABLE test_users")
	_, err = db.ExecContext(ctx, "CREATE TABLE test_users (id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY, name VARCHAR2(255))")
	if err != nil {
		log.Fatalf("create table failed: %v", err)
	}

	// Test Insert with Returning
	var id int64
	// go-ora might need sql.Out for the returning parameter
	// But let's try QueryRow first
	err = db.QueryRowContext(ctx, "INSERT INTO test_users (name) VALUES (:1) RETURNING id INTO :2", "Alice", &id).Scan()
	if err != nil {
		fmt.Printf("Insert with direct pointer failed: %v\n", err)
	} else {
		fmt.Printf("Insert with direct pointer worked: ID=%d\n", id)
	}

	// Try Named and Out
	err = db.QueryRowContext(ctx, "INSERT INTO test_users (name) VALUES (:1) RETURNING id INTO :id", "David", sql.Named("id", sql.Out{Dest: &id})).Scan()
	if err != nil {
		fmt.Printf("Insert with Named/Out failed: %v\n", err)
	} else {
		fmt.Printf("Insert with Named/Out worked: ID=%d\n", id)
	}
}
