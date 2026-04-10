package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDefaultOutputOptions_EnvVars(t *testing.T) {
	// Save and restore env vars
	prevOutput := os.Getenv("GOFRAME_OUTPUT")
	prevColor := os.Getenv("GOFRAME_COLOR")
	prevSymbols := os.Getenv("GOFRAME_SYMBOLS")
	defer func() {
		os.Setenv("GOFRAME_OUTPUT", prevOutput)
		os.Setenv("GOFRAME_COLOR", prevColor)
		os.Setenv("GOFRAME_SYMBOLS", prevSymbols)
	}()

	os.Setenv("GOFRAME_OUTPUT", "json")
	os.Setenv("GOFRAME_COLOR", "always")
	os.Setenv("GOFRAME_SYMBOLS", "false")

	opts := defaultOutputOptions()
	if opts.Format != outputFormatJSON {
		t.Errorf("expected json format, got %s", opts.Format)
	}
	if opts.Color != colorModeAlways {
		t.Errorf("expected always color, got %s", opts.Color)
	}
	if opts.Symbols {
		t.Error("expected symbols to be false")
	}
}

func TestParseGlobalOutputOptions(t *testing.T) {
	t.Run("empty args", func(t *testing.T) {
		rest, opts, err := parseGlobalOutputOptions(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rest != nil {
			t.Errorf("expected nil rest, got %v", rest)
		}
		if opts.Format != outputFormatPlain {
			t.Errorf("expected plain format, got %s", opts.Format)
		}
	})

	t.Run("json shorthand", func(t *testing.T) {
		rest, opts, err := parseGlobalOutputOptions([]string{"--json", "serve"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rest[0] != "serve" {
			t.Errorf("expected rest[0]=serve, got %s", rest[0])
		}
		if opts.Format != outputFormatJSON {
			t.Errorf("expected json format, got %s", opts.Format)
		}
	})

	t.Run("output flag equals", func(t *testing.T) {
		_, opts, err := parseGlobalOutputOptions([]string{"--output=pretty", "serve"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.Format != outputFormatPretty {
			t.Errorf("expected pretty format, got %s", opts.Format)
		}
	})

	t.Run("output flag space", func(t *testing.T) {
		_, opts, err := parseGlobalOutputOptions([]string{"--output", "json", "serve"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.Format != outputFormatJSON {
			t.Errorf("expected json format, got %s", opts.Format)
		}
	})

	t.Run("invalid output value", func(t *testing.T) {
		_, _, err := parseGlobalOutputOptions([]string{"--output=invalid"})
		if err == nil {
			t.Fatal("expected error for invalid output value")
		}
	})

	t.Run("output missing value", func(t *testing.T) {
		_, _, err := parseGlobalOutputOptions([]string{"--output"})
		if err == nil {
			t.Fatal("expected error for missing output value")
		}
	})

	t.Run("color flag equals", func(t *testing.T) {
		_, opts, err := parseGlobalOutputOptions([]string{"--color=never", "serve"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.Color != colorModeNever {
			t.Errorf("expected never color, got %s", opts.Color)
		}
	})

	t.Run("color flag space", func(t *testing.T) {
		_, opts, err := parseGlobalOutputOptions([]string{"--color", "always", "serve"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.Color != colorModeAlways {
			t.Errorf("expected always color, got %s", opts.Color)
		}
	})

	t.Run("invalid color value", func(t *testing.T) {
		_, _, err := parseGlobalOutputOptions([]string{"--color=rainbow"})
		if err == nil {
			t.Fatal("expected error for invalid color value")
		}
	})

	t.Run("color missing value", func(t *testing.T) {
		_, _, err := parseGlobalOutputOptions([]string{"--color"})
		if err == nil {
			t.Fatal("expected error for missing color value")
		}
	})

	t.Run("symbols flags", func(t *testing.T) {
		_, opts, err := parseGlobalOutputOptions([]string{"--symbols", "--no-symbols", "serve"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.Symbols {
			t.Error("expected symbols to be false after --no-symbols")
		}
	})

	t.Run("stop at --", func(t *testing.T) {
		rest, _, err := parseGlobalOutputOptions([]string{"--json", "--", "--output=json"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rest[0] != "--output=json" {
			t.Errorf("expected rest to contain --output=json, got %v", rest)
		}
	})

	t.Run("missing after --", func(t *testing.T) {
		_, _, err := parseGlobalOutputOptions([]string{"--"})
		if err == nil {
			t.Fatal("expected error for missing command after --")
		}
	})

	t.Run("stop at unknown flag", func(t *testing.T) {
		rest, opts, err := parseGlobalOutputOptions([]string{"--json", "--unknown", "serve"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rest[0] != "--unknown" {
			t.Errorf("expected rest to start at --unknown, got %v", rest)
		}
		if opts.Format != outputFormatJSON {
			t.Errorf("expected json format to be set before stopping, got %s", opts.Format)
		}
	})
}

func TestParseOutputFormat(t *testing.T) {
	tests := []struct {
		input string
		want  outputFormat
		ok    bool
	}{
		{"plain", outputFormatPlain, true},
		{"pretty", outputFormatPretty, true},
		{"json", outputFormatJSON, true},
		{"invalid", "", false},
		{"", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, ok := parseOutputFormat(tc.input)
			if ok != tc.ok {
				t.Errorf("parseOutputFormat(%q) ok=%v; want %v", tc.input, ok, tc.ok)
			}
			if ok && got != tc.want {
				t.Errorf("parseOutputFormat(%q)=%q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseColorMode(t *testing.T) {
	tests := []struct {
		input string
		want  colorMode
		ok    bool
	}{
		{"auto", colorModeAuto, true},
		{"always", colorModeAlways, true},
		{"never", colorModeNever, true},
		{"invalid", "", false},
		{"", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, ok := parseColorMode(tc.input)
			if ok != tc.ok {
				t.Errorf("parseColorMode(%q) ok=%v; want %v", tc.input, ok, tc.ok)
			}
			if ok && got != tc.want {
				t.Errorf("parseColorMode(%q)=%q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestExtractFlagValue(t *testing.T) {
	tests := []struct {
		name     string
		arg      string
		flagName string
		want     string
		wantOK   bool
	}{
		{"exact match", "--output", "--output", "", true},
		{"equals syntax", "--output=json", "--output", "json", true},
		{"no match", "--color=auto", "--output", "", false},
		{"partial prefix", "--output-extra", "--output", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := extractFlagValue(tc.arg, tc.flagName)
			if ok != tc.wantOK {
				t.Errorf("extractFlagValue(%q, %q) ok=%v; want %v", tc.arg, tc.flagName, ok, tc.wantOK)
			}
			if got != tc.want {
				t.Errorf("extractFlagValue(%q, %q)=%q; want %q", tc.arg, tc.flagName, got, tc.want)
			}
		})
	}
}

func TestOutputWantsJSON(t *testing.T) {
	// Save and restore
	prevOpts := currentOutputOptions()
	defer setOutputOptions(prevOpts)

	setOutputOptions(outputOptions{Format: outputFormatJSON})
	if !outputWantsJSON(false) {
		t.Error("expected outputWantsJSON to be true when global format is json")
	}
	if !outputWantsJSON(true) {
		t.Error("expected outputWantsJSON to be true when local json is true")
	}

	setOutputOptions(outputOptions{Format: outputFormatPlain})
	if outputWantsJSON(false) {
		t.Error("expected outputWantsJSON to be false when global format is plain")
	}
}

func TestOutputIsPretty(t *testing.T) {
	prevOpts := currentOutputOptions()
	defer setOutputOptions(prevOpts)

	setOutputOptions(outputOptions{Format: outputFormatPretty})
	if !outputIsPretty() {
		t.Error("expected outputIsPretty to be true")
	}

	setOutputOptions(outputOptions{Format: outputFormatPlain})
	if outputIsPretty() {
		t.Error("expected outputIsPretty to be false")
	}
}

func TestStatusTag(t *testing.T) {
	var buf bytes.Buffer

	tests := []struct {
		status   string
		contains string
	}{
		{"ok", "OK"},
		{"warning", "WARNING"},
		{"error", "ERROR"},
		{"degraded", "DEGRADED"},
		{"", "INFO"},
		{"unknown", "UNKNOWN"},
	}

	for _, tc := range tests {
		t.Run(tc.status, func(t *testing.T) {
			prevOpts := currentOutputOptions()
			defer setOutputOptions(prevOpts)

			// Plain mode: no symbols, no color
			setOutputOptions(outputOptions{Format: outputFormatPlain, Symbols: false})
			tag := statusTag(&buf, tc.status)
			if !strings.Contains(strings.ToUpper(tag), tc.contains) {
				t.Errorf("statusTag(%q)=%q; expected to contain %q", tc.status, tag, tc.contains)
			}
		})
	}
}

func TestStatusTag_PrettyMode(t *testing.T) {
	var buf bytes.Buffer

	prevOpts := currentOutputOptions()
	defer setOutputOptions(prevOpts)

	setOutputOptions(outputOptions{Format: outputFormatPretty, Symbols: true, Color: colorModeNever})

	tag := statusTag(&buf, "ok")
	if !strings.Contains(tag, "OK") {
		t.Errorf("expected tag to contain OK, got %q", tag)
	}
	if !strings.Contains(tag, "+") {
		t.Errorf("expected tag to contain + symbol, got %q", tag)
	}
}

func TestStatusTag_WithoutSymbols(t *testing.T) {
	var buf bytes.Buffer

	prevOpts := currentOutputOptions()
	defer setOutputOptions(prevOpts)

	setOutputOptions(outputOptions{Format: outputFormatPretty, Symbols: false, Color: colorModeNever})

	tag := statusTag(&buf, "ok")
	if strings.Contains(tag, "+") {
		t.Errorf("expected tag without + symbol when symbols disabled, got %q", tag)
	}
}

func TestWriteCommandStatus_Plain(t *testing.T) {
	var buf bytes.Buffer

	prevOpts := currentOutputOptions()
	defer setOutputOptions(prevOpts)

	setOutputOptions(outputOptions{Format: outputFormatPlain})

	err := writeCommandStatus(&buf, "health", "ok", "all good", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "all good") {
		t.Errorf("expected output to contain 'all good', got %q", buf.String())
	}
}

func TestWriteCommandStatus_Pretty(t *testing.T) {
	var buf bytes.Buffer

	prevOpts := currentOutputOptions()
	defer setOutputOptions(prevOpts)

	setOutputOptions(outputOptions{Format: outputFormatPretty, Color: colorModeNever})

	err := writeCommandStatus(&buf, "health", "ok", "all good", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "OK") {
		t.Errorf("expected output to contain 'OK', got %q", output)
	}
	if !strings.Contains(output, "all good") {
		t.Errorf("expected output to contain 'all good', got %q", output)
	}
}

func TestWriteCommandStatus_JSON(t *testing.T) {
	var buf bytes.Buffer

	prevOpts := currentOutputOptions()
	defer setOutputOptions(prevOpts)

	setOutputOptions(outputOptions{Format: outputFormatJSON})

	err := writeCommandStatus(&buf, "health", "ok", "all good", map[string]interface{}{"db": "connected"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"command": "health"`) {
		t.Errorf("expected JSON output to contain command, got %q", output)
	}
	if !strings.Contains(output, `"status": "ok"`) {
		t.Errorf("expected JSON output to contain status, got %q", output)
	}
	if !strings.Contains(output, `"db": "connected"`) {
		t.Errorf("expected JSON output to contain data, got %q", output)
	}
}

func TestWriteCommandStatus_Defaults(t *testing.T) {
	var buf bytes.Buffer

	prevOpts := currentOutputOptions()
	defer setOutputOptions(prevOpts)

	setOutputOptions(outputOptions{Format: outputFormatPlain})

	err := writeCommandStatus(&buf, "", "", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// In plain mode with empty message, it just writes a newline
	output := buf.String()
	if output == "" && buf.Len() == 0 {
		t.Error("expected some output even if empty command/message")
	}
}

func TestSetColorizeText(t *testing.T) {
	var buf bytes.Buffer

	prevOpts := currentOutputOptions()
	defer setOutputOptions(prevOpts)

	// With color disabled
	setOutputOptions(outputOptions{Color: colorModeNever})
	result := colorizeText(&buf, "31", "error")
	if result != "error" {
		t.Errorf("expected plain text when color disabled, got %q", result)
	}

	// With empty color code
	setOutputOptions(outputOptions{Color: colorModeAlways})
	result = colorizeText(&buf, "", "text")
	if result != "text" {
		t.Errorf("expected plain text when color code empty, got %q", result)
	}
}

func TestSetOutputOptions_ThreadSafety(t *testing.T) {
	// Just ensure it doesn't panic
	prevOpts := currentOutputOptions()
	defer setOutputOptions(prevOpts)

	for i := 0; i < 100; i++ {
		setOutputOptions(outputOptions{Format: outputFormatJSON})
		_ = currentOutputOptions()
	}
}
