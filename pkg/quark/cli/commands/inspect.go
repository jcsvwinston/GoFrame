package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jcsvwinston/GoFrame/pkg/quark/cli/internal/db"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	inspectFormat string
	inspectModel  string
)

func init() {
	inspectCmd.AddCommand(inspectSchemaCmd)
	inspectCmd.AddCommand(inspectTableCmd)
	inspectCmd.AddCommand(inspectSQLCmd)

	inspectCmd.PersistentFlags().StringVar(&inspectFormat, "format", "table", "Output format (table|json|yaml)")
	inspectSQLCmd.Flags().StringVar(&inspectModel, "model", "", "Model name")

	rootCmd.AddCommand(inspectCmd)
}

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Database introspection tools",
}

var inspectSchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Show full database schema",
	Run: func(cmd *cobra.Command, args []string) {
		runInspectSchema()
	},
}

var inspectTableCmd = &cobra.Command{
	Use:   "table <name>",
	Short: "Show structure of a specific table",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runInspectTable(args[0])
	},
}

var inspectSQLCmd = &cobra.Command{
	Use:   "sql",
	Short: "Show generated SQL for a model",
	Run: func(cmd *cobra.Command, args []string) {
		runInspectSQL()
	},
}

func runInspectSchema() {
	color.Yellow("Inspecting full schema...")
	// Logic to list all tables and their columns
}

func runInspectTable(name string) {
	client, err := db.GetQuarkClient()
	if err != nil {
		color.Red("Error: %v", err)
		return
	}

	info, err := db.GetTableInfo(client.Raw(), client.Dialect().Name(), name)
	if err != nil {
		color.Red("Error introspecting table %s: %v", name, err)
		return
	}

	fmt.Printf("Table: %s\n", name)
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Column", "Type", "Nullable", "PK", "Auto", "Default"})

	for _, col := range info.Columns {
		table.Append([]string{
			col.Name,
			col.Type,
			fmt.Sprintf("%v", col.IsNullable),
			fmt.Sprintf("%v", col.IsPK),
			fmt.Sprintf("%v", col.IsAuto),
			col.Default.String,
		})
	}
	table.Render()
}

func runInspectSQL() {
	color.Yellow("SQL generation inspection not yet implemented.")
}
