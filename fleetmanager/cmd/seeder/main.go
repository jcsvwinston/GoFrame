package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func seedDB(path string) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("failed to open %s: %v", path, err)
	}
	defer db.Close()

	rand.Seed(time.Now().UnixNano())

	// Clear existing data
	tables := []string{
		"telemetries", "alerts", "maintenance_tasks", "sensors",
		"assets", "devices", "drivers", "geofences",
		"trips", "fleets", "organizations",
	}
	for _, t := range tables {
		_, _ = db.Exec("DELETE FROM " + t)
		_, _ = db.Exec("DELETE FROM sqlite_sequence WHERE name='" + t + "'")
	}

	fmt.Printf("Seeding %s with hundreds of records...\n", path)

	// 1. Organizations (10)
	orgs := []int64{}
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("Organization %02d (%s)", i, path)
		slug := fmt.Sprintf("org-%02d-%s", i, path)
		res, err := db.Exec("INSERT INTO organizations (name, slug, created_at, updated_at) VALUES (?, ?, ?, ?)",
			name, slug, time.Now(), time.Now())
		if err != nil {
			log.Fatalf("org %d: %v", i, err)
		}
		id, _ := res.LastInsertId()
		orgs = append(orgs, id)
	}

	// 2. Fleets (50)
	fleets := []int64{}
	for i := 1; i <= 50; i++ {
		orgID := orgs[rand.Intn(len(orgs))]
		name := fmt.Sprintf("Fleet %02d (%s)", i, path)
		res, err := db.Exec("INSERT INTO fleets (organization_id, name, created_at, updated_at) VALUES (?, ?, ?, ?)",
			orgID, name, time.Now(), time.Now())
		if err != nil {
			log.Fatalf("fleet %d: %v", i, err)
		}
		id, _ := res.LastInsertId()
		fleets = append(fleets, id)
	}

	// 3. Devices (200)
	devices := []int64{}
	for i := 1; i <= 200; i++ {
		fleetID := fleets[rand.Intn(len(fleets))]
		name := fmt.Sprintf("Device %03d (%s)", i, path)
		serial := fmt.Sprintf("SN-%08d-%s", rand.Intn(100000000), path)
		res, err := db.Exec("INSERT INTO devices (fleet_id, name, serial, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			fleetID, name, serial, time.Now(), time.Now())
		if err != nil {
			log.Fatalf("device %d: %v", i, err)
		}
		id, _ := res.LastInsertId()
		devices = append(devices, id)
	}

	// 4. Sensors (500)
	sensors := []int64{}
	sensorTypes := []string{"gps", "temp", "fuel", "engine", "humidity"}
	for i := 1; i <= 500; i++ {
		deviceID := devices[rand.Intn(len(devices))]
		name := fmt.Sprintf("Sensor %03d (%s)", i, path)
		sType := sensorTypes[rand.Intn(len(sensorTypes))]
		res, err := db.Exec("INSERT INTO sensors (device_id, name, type, unit, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
			deviceID, name, sType, "unit", "online", time.Now(), time.Now())
		if err != nil {
			log.Fatalf("sensor %d: %v", i, err)
		}
		id, _ := res.LastInsertId()
		sensors = append(sensors, id)
	}

	// 5. Assets (150)
	assets := []int64{}
	assetTypes := []string{"truck", "van", "car", "trailer"}
	for i := 1; i <= 150; i++ {
		fleetID := fleets[rand.Intn(len(fleets))]
		name := fmt.Sprintf("Asset %03d (%s)", i, path)
		vin := fmt.Sprintf("VIN-%012d-%s", rand.Intn(1000000000000), path)
		res, err := db.Exec("INSERT INTO assets (fleet_id, name, type, vin, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			fleetID, name, assetTypes[rand.Intn(len(assetTypes))], vin, time.Now(), time.Now())
		if err != nil {
			log.Fatalf("asset %d: %v", i, err)
		}
		id, _ := res.LastInsertId()
		assets = append(assets, id)
	}

	// 6. Drivers (100)
	drivers := []int64{}
	for i := 1; i <= 100; i++ {
		name := fmt.Sprintf("Driver %03d (%s)", i, path)
		license := fmt.Sprintf("LIC-%06d-%s", rand.Intn(1000000), path)
		res, err := db.Exec("INSERT INTO drivers (name, license_no, phone, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			name, license, fmt.Sprintf("555-%04d", i), "active", time.Now(), time.Now())
		if err != nil {
			log.Fatalf("driver %d: %v", i, err)
		}
		id, _ := res.LastInsertId()
		drivers = append(drivers, id)
	}

	// 7. Telemetries (2000)
	for i := 1; i <= 2000; i++ {
		sensorID := sensors[rand.Intn(len(sensors))]
		value := rand.Float64() * 100
		_, err := db.Exec("INSERT INTO telemetries (sensor_id, value, timestamp, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			sensorID, value, time.Now().Add(time.Duration(-rand.Intn(10000))*time.Minute), time.Now(), time.Now())
		if err != nil {
			log.Fatalf("telemetries %d: %v", i, err)
		}
	}

	// 8. MaintenanceTasks (300)
	for i := 1; i <= 300; i++ {
		deviceID := devices[rand.Intn(len(devices))]
		desc := fmt.Sprintf("Maintenance task %d (%s)", i, path)
		status := "pending"
		if rand.Float64() > 0.5 {
			status = "completed"
		}
		_, err := db.Exec("INSERT INTO maintenance_tasks (device_id, description, status, due_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			deviceID, desc, status, time.Now().AddDate(0, 0, rand.Intn(30)), time.Now(), time.Now())
		if err != nil {
			log.Fatalf("task %d: %v", i, err)
		}
	}

	// 9. Alerts (1000)
	severities := []string{"low", "medium", "high", "critical"}
	alertTypes := []string{"speeding", "geofence", "maintenance", "sensor"}
	for i := 1; i <= 1000; i++ {
		fleetID := fleets[rand.Intn(len(fleets))]
		severity := severities[rand.Intn(len(severities))]
		aType := alertTypes[rand.Intn(len(alertTypes))]
		msg := fmt.Sprintf("Alert %04d: %s issue (%s)", i, aType, path)
		resolved := 0
		if rand.Float64() > 0.7 {
			resolved = 1
		}
		_, err := db.Exec("INSERT INTO alerts (fleet_id, type, severity, message, resolved, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
			fleetID, aType, severity, msg, resolved, time.Now(), time.Now())
		if err != nil {
			log.Fatalf("alert %d: %v", i, err)
		}
	}

	// 10. Geofences (50)
	for i := 1; i <= 50; i++ {
		fleetID := fleets[rand.Intn(len(fleets))]
		name := fmt.Sprintf("Geofence %02d (%s)", i, path)
		area := "POLYGON((...))"
		_, err := db.Exec("INSERT INTO geofences (fleet_id, name, area, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			fleetID, name, area, time.Now(), time.Now())
		if err != nil {
			log.Fatalf("geofence %d: %v", i, err)
		}
	}

	// 11. Trips (300)
	for i := 1; i <= 300; i++ {
		assetID := assets[rand.Intn(len(assets))]
		driverID := drivers[rand.Intn(len(drivers))]
		start := time.Now().Add(time.Duration(-rand.Intn(1000))*time.Hour)
		end := start.Add(time.Duration(rand.Intn(10))*time.Hour)
		dist := rand.Float64() * 500
		_, err := db.Exec("INSERT INTO trips (asset_id, driver_id, start_time, end_time, distance, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
			assetID, driverID, start, end, dist, time.Now(), time.Now())
		if err != nil {
			log.Fatalf("trip %d: %v", i, err)
		}
	}

	fmt.Printf("Successfully seeded %s with 11 entities!\n", path)
}

func main() {
	seedDB("app.db")
	seedDB("secondary.db")
}

