package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./fleetmanager/app.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Seed random
	rand.Seed(time.Now().UnixNano())

	// Tables to seed
	tables := []string{
		"organizations", "fleets", "devices", "sensors", 
		"telemetries", "drivers", "assets", "maintenance_tasks", 
		"alerts", "geofences", "trips",
	}

	for _, table := range tables {
		fmt.Printf("Seeding table: %s...\n", table)
		count := 250 // Generate 250 records per table
		for i := 0; i < count; i++ {
			switch table {
			case "organizations":
				_, _ = db.Exec("INSERT INTO organizations (created_at, updated_at, name, slug) VALUES (?, ?, ?, ?)",
					time.Now(), time.Now(), fmt.Sprintf("Org %d", i), fmt.Sprintf("org-%d", i))
			case "fleets":
				_, _ = db.Exec("INSERT INTO fleets (created_at, updated_at, organization_id, name) VALUES (?, ?, ?, ?)",
					time.Now(), time.Now(), 1, fmt.Sprintf("Fleet %d", i))
			case "devices":
				_, _ = db.Exec("INSERT INTO devices (created_at, updated_at, fleet_id, name, serial) VALUES (?, ?, ?, ?, ?)",
					time.Now(), time.Now(), 1, fmt.Sprintf("Device %d", i), fmt.Sprintf("SN-%d-%d", i, rand.Intn(100000)))
			case "sensors":
				_, _ = db.Exec("INSERT INTO sensors (created_at, updated_at, device_id, name, type, unit, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
					time.Now(), time.Now(), 1, fmt.Sprintf("Sensor %d", i), "analog", "units", "online")
			case "telemetries":
				_, _ = db.Exec("INSERT INTO telemetries (sensor_id, value, timestamp) VALUES (?, ?, ?)",
					rand.Intn(10)+1, rand.Float64()*100, time.Now())
			case "drivers":
				_, _ = db.Exec("INSERT INTO drivers (created_at, updated_at, name, license_no, phone, status) VALUES (?, ?, ?, ?, ?, ?)",
					time.Now(), time.Now(), fmt.Sprintf("Driver %d", i), fmt.Sprintf("LIC-%d", i), "+12345", "active")
			case "assets":
				_, _ = db.Exec("INSERT INTO assets (created_at, updated_at, fleet_id, name, type, vin) VALUES (?, ?, ?, ?, ?, ?)",
					time.Now(), time.Now(), 1, fmt.Sprintf("Asset %d", i), "Truck", fmt.Sprintf("VIN-%d", i))
			case "maintenance_tasks":
				_, _ = db.Exec("INSERT INTO maintenance_tasks (created_at, updated_at, device_id, description, status, due_date) VALUES (?, ?, ?, ?, ?, ?)",
					time.Now(), time.Now(), 1, "Periodic check", "pending", time.Now().AddDate(0, 0, 7))
			case "alerts":
				_, _ = db.Exec("INSERT INTO alerts (created_at, updated_at, fleet_id, type, severity, message) VALUES (?, ?, ?, ?, ?, ?)",
					time.Now(), time.Now(), 1, "Warning", "low", "System check required")
			case "geofences":
				_, _ = db.Exec("INSERT INTO geofences (created_at, updated_at, fleet_id, name, area) VALUES (?, ?, ?, ?, ?)",
					time.Now(), time.Now(), 1, fmt.Sprintf("Zone %d", i), "POLYGON(...)")
			case "trips":
				_, _ = db.Exec("INSERT INTO trips (created_at, updated_at, asset_id, driver_id, start_time, distance) VALUES (?, ?, ?, ?, ?, ?)",
					time.Now(), time.Now(), 1, 1, time.Now(), rand.Float64()*1000)
			}
		}
	}

	fmt.Println("Seeding completed!")
}
