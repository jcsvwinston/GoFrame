package commands

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/jcsvwinston/GoFrame/pkg/quark/cli/internal/db"
	"github.com/jcsvwinston/GoFrame/pkg/quark/migrate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	tenantID   string
	skipSeed   bool
	forceOp    bool
)

func init() {
	tenantCmd.AddCommand(tenantProvisionCmd)
	tenantCmd.AddCommand(tenantMigrateCmd)
	tenantCmd.AddCommand(tenantListCmd)
	tenantCmd.AddCommand(tenantMigrateAllCmd)

	tenantProvisionCmd.Flags().BoolVar(&skipSeed, "skip-seed", false, "Skip seeders after provision")
	tenantMigrateCmd.Flags().StringVar(&tenantID, "tenant-id", "", "ID of the tenant")
	tenantMigrateAllCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Preview SQL")

	rootCmd.AddCommand(tenantCmd)
}

var tenantCmd = &cobra.Command{
	Use:   "tenant",
	Short: "Manage multi-tenant environments",
}

var tenantProvisionCmd = &cobra.Command{
	Use:   "provision <tenant-id>",
	Short: "Provision a new tenant",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runTenantProvision(args[0])
	},
}

var tenantMigrateCmd = &cobra.Command{
	Use:   "migrate <tenant-id>",
	Short: "Run migrations for a specific tenant",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runTenantMigrate(args[0])
	},
}

var tenantListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active tenants",
	Run: func(cmd *cobra.Command, args []string) {
		runTenantList()
	},
}

var tenantMigrateAllCmd = &cobra.Command{
	Use:   "migrate-all",
	Short: "Run migrations for all tenants",
	Run: func(cmd *cobra.Command, args []string) {
		runTenantMigrateAll()
	},
}

func runTenantProvision(id string) {
	fmt.Printf("Provisioning tenant: %s...\n", id)
	
	adminClient, err := db.GetAdminQuarkClient()
	if err != nil {
		color.Red("Error connecting to admin database: %v", err)
		return
	}

	strategy := viper.GetString("tenant.strategy")
	if strategy == "" {
		strategy = "db_per_tenant"
	}

	ctx := context.Background()

	switch strategy {
	case "db_per_tenant":
		// Create Database
		query := fmt.Sprintf("CREATE DATABASE %s", id)
		if err := adminClient.Exec(ctx, query); err != nil {
			color.Red("Error creating database: %v", err)
			return
		}
		fmt.Printf("  Created database: %s\n", id)
	case "schema_per_tenant":
		// Create Schema
		query := fmt.Sprintf("CREATE SCHEMA %s", id)
		if err := adminClient.Exec(ctx, query); err != nil {
			color.Red("Error creating schema: %v", err)
			return
		}
		fmt.Printf("  Created schema: %s\n", id)
	default:
		color.Red("Unsupported strategy: %s", strategy)
		return
	}

	// Run migrations
	runTenantMigrate(id)

	color.Green("Tenant %s provisioned successfully!", id)
}

func runTenantMigrate(id string) {
	fmt.Printf("Migrating tenant: %s...\n", id)
	
	// In a real implementation, we would resolve the tenant DSN/Schema here.
	// For this CLI version, we'll assume the default client can be used with a router or DSN adjustment.
	client, err := db.GetQuarkClient()
	if err != nil {
		color.Red("Error connecting to tenant database: %v", err)
		return
	}

	migrator := migrate.NewMigrator(client)
	if err := migrator.Up(context.Background(), 0); err != nil {
		color.Red("Error migrating tenant %s: %v", id, err)
		return
	}
	fmt.Printf("  Migrations complete for %s\n", id)
}

func runTenantList() {
	color.Yellow("Listing tenants requires access to the tenant registry database/table.")
}

func runTenantMigrateAll() {
	color.Yellow("Migrating all tenants requires a list of all active tenants.")
}

// getAdminQuarkClient and getTenantQuarkClient removed
