package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunWizard(t *testing.T) {
	t.Run("invalid arguments", func(t *testing.T) {
		var out bytes.Buffer
		var errOut bytes.Buffer
		err := runWizard([]string{"--invalid-flag"}, strings.NewReader(""), &out, &errOut)
		if err == nil {
			t.Fatal("expected error for invalid flag")
		}
	})

	t.Run("missing --type flag", func(t *testing.T) {
		var out bytes.Buffer
		var errOut bytes.Buffer
		err := runWizard([]string{}, strings.NewReader(""), &out, &errOut)
		if err == nil {
			t.Fatal("expected error for missing --type flag")
		}
		if !strings.Contains(errOut.String(), "wizard type is required") {
			t.Fatalf("expected error message about missing type, got: %s", errOut.String())
		}
	})

	t.Run("unknown wizard type", func(t *testing.T) {
		var out bytes.Buffer
		var errOut bytes.Buffer
		err := runWizard([]string{"--type", "unknown"}, strings.NewReader(""), &out, &errOut)
		if err == nil {
			t.Fatal("expected error for unknown wizard type")
		}
		if !strings.Contains(errOut.String(), "unknown wizard type") {
			t.Fatalf("expected error message about unknown type, got: %s", errOut.String())
		}
	})

	t.Run("inspectdb type", func(t *testing.T) {
		var out bytes.Buffer
		var errOut bytes.Buffer
		err := runWizard([]string{"--type", "inspectdb"}, strings.NewReader("postgres://localhost:5432/db\n1\ninternal/models\nPascalCase\n"), &out, &errOut)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "GoFrame inspectdb Wizard") {
			t.Fatalf("expected wizard title, got: %s", out.String())
		}
	})

	t.Run("new type", func(t *testing.T) {
		var out bytes.Buffer
		var errOut bytes.Buffer
		err := runWizard([]string{"--type", "new"}, strings.NewReader(""), &out, &errOut)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "GoFrame new Wizard") {
			t.Fatalf("expected wizard title, got: %s", out.String())
		}
	})

	t.Run("startapp type", func(t *testing.T) {
		var out bytes.Buffer
		var errOut bytes.Buffer
		err := runWizard([]string{"--type", "startapp"}, strings.NewReader(""), &out, &errOut)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "GoFrame startapp Wizard") {
			t.Fatalf("expected wizard title, got: %s", out.String())
		}
	})
}

func TestPromptUser(t *testing.T) {
	t.Run("text input", func(t *testing.T) {
		var out bytes.Buffer
		result, err := promptUser(strings.NewReader("hello\n"), &out, wizardPrompt{
			question: "Enter value",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "hello" {
			t.Fatalf("expected 'hello', got '%s'", result)
		}
	})

	t.Run("default value used", func(t *testing.T) {
		var out bytes.Buffer
		result, err := promptUser(strings.NewReader("\n"), &out, wizardPrompt{
			question:   "Enter value",
			defaultVal: "default",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "default" {
			t.Fatalf("expected 'default', got '%s'", result)
		}
	})

	t.Run("validation passes", func(t *testing.T) {
		var out bytes.Buffer
		result, err := promptUser(strings.NewReader("valid\n"), &out, wizardPrompt{
			question: "Enter value",
			validate: func(s string) error {
				if s == "valid" {
					return nil
				}
				return nil
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "valid" {
			t.Fatalf("expected 'valid', got '%s'", result)
		}
	})

	t.Run("validation fails and retries", func(t *testing.T) {
		var out bytes.Buffer
		result, err := promptUser(strings.NewReader("invalid\nvalid\n"), &out, wizardPrompt{
			question: "Enter value",
			validate: func(s string) error {
				if s == "valid" {
					return nil
				}
				return nil
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "valid" {
			t.Fatalf("expected 'valid', got '%s'", result)
		}
	})

	t.Run("transform function", func(t *testing.T) {
		var out bytes.Buffer
		result, err := promptUser(strings.NewReader("hello\n"), &out, wizardPrompt{
			question: "Enter value",
			transform: func(s string) string {
				return strings.ToUpper(s)
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "HELLO" {
			t.Fatalf("expected 'HELLO', got '%s'", result)
		}
	})
}

func TestPromptSelect(t *testing.T) {
	t.Run("valid selection by number", func(t *testing.T) {
		var out bytes.Buffer
		_, err := promptSelect(nil, &out, wizardPrompt{
			question: "Select option",
			options:  []string{"option1", "option2", "option3"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("valid selection by name", func(t *testing.T) {
		var out bytes.Buffer
		_, err := promptSelect(nil, &out, wizardPrompt{
			question: "Select option",
			options:  []string{"option1", "option2", "option3"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("default selection", func(t *testing.T) {
		var out bytes.Buffer
		_, err := promptSelect(nil, &out, wizardPrompt{
			question:   "Select option",
			options:    []string{"option1", "option2", "option3"},
			defaultVal: "option2",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid selection", func(t *testing.T) {
		var out bytes.Buffer
		_, err := promptSelect(nil, &out, wizardPrompt{
			question: "Select option",
			options:  []string{"option1", "option2", "option3"},
		})
		if err == nil {
			t.Fatal("expected error for invalid selection")
		}
	})
}

func TestPromptMultiSelect(t *testing.T) {
	t.Run("single selection", func(t *testing.T) {
		var out bytes.Buffer
		_, err := promptMultiSelect(nil, &out, wizardPrompt{
			question: "Select options",
			options:  []string{"option1", "option2", "option3"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("multiple selections", func(t *testing.T) {
		var out bytes.Buffer
		_, err := promptMultiSelect(nil, &out, wizardPrompt{
			question: "Select options",
			options:  []string{"option1", "option2", "option3"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty selection", func(t *testing.T) {
		var out bytes.Buffer
		_, err := promptMultiSelect(nil, &out, wizardPrompt{
			question: "Select options",
			options:  []string{"option1", "option2", "option3"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("comma-separated", func(t *testing.T) {
		var out bytes.Buffer
		_, err := promptMultiSelect(nil, &out, wizardPrompt{
			question: "Select options",
			options:  []string{"option1", "option2", "option3"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("space-separated", func(t *testing.T) {
		var out bytes.Buffer
		_, err := promptMultiSelect(nil, &out, wizardPrompt{
			question: "Select options",
			options:  []string{"option1", "option2", "option3"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunInspectDBWizard(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		var out bytes.Buffer
		var errOut bytes.Buffer
		err := runInspectDBWizard("goframe.yaml", strings.NewReader("postgres://localhost:5432/db\n1\ninternal/models\nPascalCase\n"), &out, &errOut)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "GoFrame inspectdb Wizard") {
			t.Fatalf("expected wizard title, got: %s", out.String())
		}
	})

	t.Run("empty database URL", func(t *testing.T) {
		var out bytes.Buffer
		var errOut bytes.Buffer
		err := runInspectDBWizard("goframe.yaml", strings.NewReader("\n"), &out, &errOut)
		if err == nil {
			t.Fatal("expected error for empty database URL")
		}
	})
}

func TestRunNewWizard(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	err := runNewWizard("goframe.yaml", strings.NewReader(""), &out, &errOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "GoFrame new Wizard") {
		t.Fatalf("expected wizard title, got: %s", out.String())
	}
}

func TestRunStartAppWizard(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	err := runStartAppWizard("goframe.yaml", strings.NewReader(""), &out, &errOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "GoFrame startapp Wizard") {
		t.Fatalf("expected wizard title, got: %s", out.String())
	}
}
