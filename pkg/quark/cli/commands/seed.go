package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	seedName string
	seedEnv  string
)

func init() {
	seedCmd.AddCommand(seedCreateCmd)
	seedCmd.AddCommand(seedRunCmd)

	seedCreateCmd.Flags().StringVar(&seedName, "name", "", "Name of the seeder")
	seedRunCmd.Flags().StringVar(&seedName, "name", "", "Name of the specific seeder to run")
	seedRunCmd.Flags().StringVar(&seedEnv, "env", "development", "Environment")

	rootCmd.AddCommand(seedCmd)
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Manage database seeders",
}

var seedCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new seeder file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runSeedCreate(args[0])
	},
}

var seedRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run seeders",
	Run: func(cmd *cobra.Command, args []string) {
		runSeedRun()
	},
}

func runSeedCreate(name string) {
	filename := fmt.Sprintf("%s_seeder.go", name)
	dir := viper.GetString("paths.seeders")
	if dir == "" {
		dir = "./seeders"
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		color.Red("Error creating seeders directory: %v", err)
		return
	}

	path := filepath.Join(dir, filename)
	
	data := struct {
		Name string
	}{
		Name: name,
	}

	tmpl, _ := template.New("seeder").Parse(seederTemplate)
	file, err := os.Create(path)
	if err != nil {
		color.Red("Error creating seeder file: %v", err)
		return
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		color.Red("Error executing template: %v", err)
		return
	}

	fmt.Printf("Created seeder: %s\n", path)
}

func runSeedRun() {
	color.Yellow("Seed execution logic depends on the specific project models. Template created, but 'run' requires manual implementation or dynamic loading.")
}
