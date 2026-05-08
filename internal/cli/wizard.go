package cli

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type wizardPrompt struct {
	question    string
	defaultVal  string
	validate    func(string) error
	transform   func(string) string
	options     []string
	multiSelect bool
}

func runWizard(args []string, _ io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("wizard", flag.ContinueOnError)
	fs.SetOutput(stderr)

	wizardType := fs.String("type", "", "Wizard type: inspectdb, new, startapp")
	configPath := fs.String("config", "", "Path to goframe config file")

	if err := fs.Parse(args); err != nil {
		return err
	}

	rest := fs.Args()
	if len(rest) > 0 {
		return fmt.Errorf("wizard does not accept positional arguments")
	}

	if *wizardType == "" {
		return fmt.Errorf("wizard type is required (use --type inspectdb|new|startapp)")
	}

	switch *wizardType {
	case "inspectdb":
		return runInspectDBWizard(*configPath, os.Stdin, stdout, stderr)
	case "new":
		return runNewWizard(*configPath, os.Stdin, stdout, stderr)
	case "startapp":
		return runStartAppWizard(*configPath, os.Stdin, stdout, stderr)
	default:
		return fmt.Errorf("unknown wizard type: %s (use inspectdb|new|startapp)", *wizardType)
	}
}

func promptUser(reader io.Reader, writer io.Writer, p wizardPrompt) (string, error) {
	scanner := bufio.NewScanner(reader)

	if len(p.options) > 0 {
		if p.multiSelect {
			return promptMultiSelect(scanner, writer, p)
		}
		return promptSelect(scanner, writer, p)
	}

	return promptUserWithScanner(scanner, writer, p)
}

func promptUserWithScanner(scanner *bufio.Scanner, writer io.Writer, p wizardPrompt) (string, error) {
	defaultStr := ""
	if p.defaultVal != "" {
		defaultStr = fmt.Sprintf(" [%s]", p.defaultVal)
	}

	fmt.Fprintf(writer, "%s%s: ", p.question, defaultStr)

	if !scanner.Scan() {
		return "", scanner.Err()
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" && p.defaultVal != "" {
		input = p.defaultVal
	}

	if p.transform != nil {
		input = p.transform(input)
	}

	if p.validate != nil {
		if err := p.validate(input); err != nil {
			fmt.Fprintf(writer, "Error: %v\n", err)
			return promptUserWithScanner(scanner, writer, p)
		}
	}

	return input, nil
}

func promptSelect(scanner *bufio.Scanner, writer io.Writer, p wizardPrompt) (string, error) {
	fmt.Fprintf(writer, "%s\n", p.question)
	for i, opt := range p.options {
		fmt.Fprintf(writer, "  [%d] %s\n", i+1, opt)
	}
	defaultStr := ""
	if p.defaultVal != "" {
		defaultStr = fmt.Sprintf(" (default: %s)", p.defaultVal)
	}
	fmt.Fprintf(writer, "Select option%s: ", defaultStr)

	if !scanner.Scan() {
		return "", scanner.Err()
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" && p.defaultVal != "" {
		input = p.defaultVal
	}

	for i, opt := range p.options {
		if fmt.Sprintf("%d", i+1) == input || strings.EqualFold(opt, input) {
			return opt, nil
		}
	}

	return "", fmt.Errorf("invalid selection: %s", input)
}

func promptMultiSelect(scanner *bufio.Scanner, writer io.Writer, p wizardPrompt) (string, error) {
	fmt.Fprintf(writer, "%s (use space to select, enter to finish)\n", p.question)
	for i, opt := range p.options {
		fmt.Fprintf(writer, "  [%d] %s\n", i+1, opt)
	}
	fmt.Fprintf(writer, "Select options (comma-separated or space-separated numbers): ")

	if !scanner.Scan() {
		return "", scanner.Err()
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return "", nil
	}

	parts := strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || r == ' '
	})

	selected := []string{}
	for _, part := range parts {
		for i, opt := range p.options {
			if fmt.Sprintf("%d", i+1) == part || strings.EqualFold(opt, part) {
				selected = append(selected, opt)
				break
			}
		}
	}

	return strings.Join(selected, ","), nil
}

func runInspectDBWizard(configPath string, stdin io.Reader, stdout, stderr io.Writer) error {
	fmt.Fprintf(stdout, "=== GoFrame inspectdb Wizard ===\n\n")

	// Step 1: Database URL
	dbURL, err := promptUser(stdin, stdout, wizardPrompt{
		question:   "Database URL",
		defaultVal: "postgres://localhost:5432/mydb",
		validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("database URL cannot be empty")
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	// Step 2: Table selection
	fmt.Fprintf(stdout, "\nFetching tables from database...\n")
	// TODO: Actually fetch tables from database
	tables := []string{"users", "posts", "comments", "tags"} // Placeholder

	selectedTables, err := promptUser(stdin, stdout, wizardPrompt{
		question:    "Select tables to import",
		options:     tables,
		multiSelect: true,
	})
	if err != nil {
		return err
	}

	// Step 3: Output package
	outputPackage, err := promptUser(stdin, stdout, wizardPrompt{
		question:   "Output package path",
		defaultVal: "internal/models",
	})
	if err != nil {
		return err
	}

	// Step 4: Naming convention
	namingConvention, err := promptUser(stdin, stdout, wizardPrompt{
		question:   "Model naming convention",
		options:    []string{"PascalCase", "snake_case", "camelCase"},
		defaultVal: "PascalCase",
	})
	if err != nil {
		return err
	}

	// Summary
	fmt.Fprintf(stdout, "\n=== Summary ===\n")
	fmt.Fprintf(stdout, "Database URL: %s\n", dbURL)
	fmt.Fprintf(stdout, "Selected tables: %s\n", selectedTables)
	fmt.Fprintf(stdout, "Output package: %s\n", outputPackage)
	fmt.Fprintf(stdout, "Naming convention: %s\n", namingConvention)

	// TODO: Actually run inspectdb with these parameters
	fmt.Fprintf(stdout, "\nWizard mode is a framework placeholder. Use: goframe inspectdb --config goframe.yaml\n")

	return nil
}

func runNewWizard(configPath string, stdin io.Reader, stdout, stderr io.Writer) error {
	fmt.Fprintf(stdout, "=== GoFrame new Wizard ===\n\n")
	fmt.Fprintf(stdout, "Wizard mode is a framework placeholder. Use: goframe new <name> --module <module>\n")
	return nil
}

func runStartAppWizard(configPath string, stdin io.Reader, stdout, stderr io.Writer) error {
	fmt.Fprintf(stdout, "=== GoFrame startapp Wizard ===\n\n")
	fmt.Fprintf(stdout, "Wizard mode is a framework placeholder. Use: goframe startapp <name> --out <path>\n")
	return nil
}
