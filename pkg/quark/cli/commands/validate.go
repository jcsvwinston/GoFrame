package commands

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	validateStrict bool
	validateModels string
)

func init() {
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "Fail if there are unmapped columns in DB")
	validateCmd.Flags().StringVar(&validateModels, "models", "", "Specific models to validate")
	rootCmd.AddCommand(validateCmd)
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate models against database schema",
	Run: func(cmd *cobra.Command, args []string) {
		runValidate()
	},
}

func runValidate() {
	color.Yellow("Model validation against DB schema is not yet implemented.")
	fmt.Println("This command will compare Go struct tags with the actual database structure.")
}
