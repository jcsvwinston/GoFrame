package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jcsvwinston/fleetmanager/internal/controllers"
	"github.com/jcsvwinston/fleetmanager/internal/models"
	"github.com/jcsvwinston/fleetmanager/internal/repositories"
	"github.com/jcsvwinston/fleetmanager/internal/services"
	"github.com/jcsvwinston/nucleus/pkg/admin"
	"github.com/jcsvwinston/nucleus/pkg/app"
	"github.com/jcsvwinston/nucleus/pkg/model"
)

func main() {
	cfg, err := app.LoadConfig("goframe.yaml")
	if err != nil {
		log.Fatal(err)
	}

	a, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := a.DB.SqlDB()
	if err != nil {
		log.Fatal(err)
	}
	if err := ensureSchema(sqlDB); err != nil {
		log.Fatal(err)
	}

	// Bootstrap Admin User
	res, err := admin.EnsureBootstrapAdminUser(context.Background(), sqlDB, admin.BootstrapAdminConfig{
		Username: cfg.AdminBootstrapUsername,
		Email:    cfg.AdminBootstrapEmail,
		Password: cfg.AdminBootstrapPassword,
	})
	if err != nil {
		log.Printf("Warning: admin bootstrap failed: %v", err)
	} else if res.Created {
		log.Printf("Admin user %q created successfully", res.Username)
	}

	// Initialize secondary database schema
	secondaryDB, ok := a.DBs["secondary"]
	if ok && secondaryDB != nil {
		secSQL, err := secondaryDB.SqlDB()
		if err == nil {
			_ = ensureSchema(secSQL)
		}
	}

	// Register Admin Models
	a.RegisterModel(&models.Organization{}, model.ModelConfig{Icon: "office-building"})
	a.RegisterModel(&models.Fleet{}, model.ModelConfig{Icon: "collection"})
	a.RegisterModel(&models.Device{}, model.ModelConfig{Icon: "chip"})
	a.RegisterModel(&models.Sensor{}, model.ModelConfig{Icon: "lightning-bolt"})
	a.RegisterModel(&models.Telemetry{}, model.ModelConfig{Icon: "chart-bar"})
	a.RegisterModel(&models.MaintenanceTask{}, model.ModelConfig{Icon: "clipboard-list"})
	a.RegisterModel(&models.Driver{}, model.ModelConfig{Icon: "user-group"})
	a.RegisterModel(&models.Asset{}, model.ModelConfig{Icon: "truck"})
	a.RegisterModel(&models.Trip{}, model.ModelConfig{Icon: "map", DatabaseAlias: "secondary"})
	a.RegisterModel(&models.Geofence{}, model.ModelConfig{Icon: "view-grid", DatabaseAlias: "secondary"})
	a.RegisterModel(&models.Alert{}, model.ModelConfig{Icon: "bell", DatabaseAlias: "secondary"})

	// Wire All 11 API Handlers
	registerAPI(a, sqlDB)

	// Serve React SPA
	staticDir := "internal/web/static"
	a.Router.Mount("/assets", http.FileServer(http.Dir(filepath.Join(staticDir, "assets"))))
	a.Router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(staticDir, r.URL.Path)
		if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/admin") {
			return
		}
		if _, err := os.Stat(path); err == nil && !strings.HasSuffix(path, "/") {
			http.ServeFile(w, r, path)
			return
		}
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	}))

	log.Println("fleetmanager running:")
	log.Printf("  web:   http://localhost:%d/\n", cfg.Port)
	log.Printf("  admin: http://localhost:%d/admin\n", cfg.Port)
	log.Fatal(a.Run(context.Background()))
}

func registerAPI(a *app.App, db *sql.DB) {
	// Organizations
	orgRepo := repositories.NewOrganizationRepository(db)
	orgService := services.NewOrganizationService(orgRepo)
	controllers.NewOrganizationHandler(orgService).Mount(a.Router.Mux)

	// Fleets
	fleetRepo := repositories.NewFleetRepository(db)
	fleetService := services.NewFleetService(fleetRepo)
	controllers.NewFleetHandler(fleetService).Mount(a.Router.Mux)

	// Devices
	deviceRepo := repositories.NewDeviceRepository(db)
	deviceService := services.NewDeviceService(deviceRepo)
	controllers.NewDeviceHandler(deviceService).Mount(a.Router.Mux)

	// Sensors
	sensorRepo := repositories.NewSensorRepository(db)
	sensorService := services.NewSensorService(sensorRepo)
	controllers.NewSensorHandler(sensorService).Mount(a.Router.Mux)

	// Telemetry
	telemetryRepo := repositories.NewTelemetryRepository(db)
	telemetryService := services.NewTelemetryService(telemetryRepo)
	controllers.NewTelemetryHandler(telemetryService).Mount(a.Router.Mux)

	// Maintenance
	maintRepo := repositories.NewMaintenanceTaskRepository(db)
	maintService := services.NewMaintenanceTaskService(maintRepo)
	controllers.NewMaintenanceTaskHandler(maintService).Mount(a.Router.Mux)

	// Drivers
	driverRepo := repositories.NewDriverRepository(db)
	driverService := services.NewDriverService(driverRepo)
	controllers.NewDriverHandler(driverService).Mount(a.Router.Mux)

	// Assets
	assetRepo := repositories.NewAssetRepository(db)
	assetService := services.NewAssetService(assetRepo)
	controllers.NewAssetHandler(assetService).Mount(a.Router.Mux)

	// Trips
	tripRepo := repositories.NewTripRepository(db)
	tripService := services.NewTripService(tripRepo)
	controllers.NewTripHandler(tripService).Mount(a.Router.Mux)

	// Geofences
	geoRepo := repositories.NewGeofenceRepository(db)
	geoService := services.NewGeofenceService(geoRepo)
	controllers.NewGeofenceHandler(geoService).Mount(a.Router.Mux)

	// Alerts
	alertRepo := repositories.NewAlertRepository(db)
	alertService := services.NewAlertService(alertRepo)
	controllers.NewAlertHandler(alertService).Mount(a.Router.Mux)
}

func ensureSchema(sqlDB *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS organizations (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, 
			name TEXT NOT NULL, slug TEXT NOT NULL UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS fleets (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, 
			organization_id INTEGER NOT NULL, name TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS devices (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, 
			fleet_id INTEGER NOT NULL, name TEXT NOT NULL, serial TEXT NOT NULL UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS sensors (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, 
			device_id INTEGER NOT NULL, name TEXT NOT NULL, type TEXT NOT NULL, 
			unit TEXT, status TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS telemetries (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME,
			sensor_id INTEGER NOT NULL, value REAL NOT NULL, 
			timestamp DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS drivers (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, 
			name TEXT NOT NULL, license_no TEXT NOT NULL UNIQUE, phone TEXT, status TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS assets (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, 
			fleet_id INTEGER NOT NULL, name TEXT NOT NULL, type TEXT NOT NULL, 
			vin TEXT NOT NULL UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS maintenance_tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, 
			device_id INTEGER NOT NULL, description TEXT NOT NULL, 
			status TEXT NOT NULL, due_date DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, 
			fleet_id INTEGER NOT NULL, type TEXT NOT NULL, severity TEXT NOT NULL, 
			message TEXT NOT NULL, resolved INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS geofences (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, 
			fleet_id INTEGER NOT NULL, name TEXT NOT NULL, area TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS trips (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, 
			asset_id INTEGER NOT NULL, driver_id INTEGER NOT NULL, 
			start_time DATETIME NOT NULL, end_time DATETIME, distance REAL
		)`,
	}
	for _, q := range queries {
		if _, err := sqlDB.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
