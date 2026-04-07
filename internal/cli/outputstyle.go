package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type outputFormat string

const (
	outputFormatPlain  outputFormat = "plain"
	outputFormatPretty outputFormat = "pretty"
	outputFormatJSON   outputFormat = "json"
)

type colorMode string

const (
	colorModeAuto   colorMode = "auto"
	colorModeAlways colorMode = "always"
	colorModeNever  colorMode = "never"
)

type outputOptions struct {
	Format  outputFormat
	Color   colorMode
	Symbols bool
}

var (
	outputOptionsMu sync.RWMutex
	outputOptionsV  = defaultOutputOptions()
)

func defaultOutputOptions() outputOptions {
	opts := outputOptions{
		Format:  outputFormatPlain,
		Color:   colorModeAuto,
		Symbols: true,
	}

	if raw := strings.ToLower(strings.TrimSpace(os.Getenv("GOFRAME_OUTPUT"))); raw != "" {
		if f, ok := parseOutputFormat(raw); ok {
			opts.Format = f
		}
	}
	if raw := strings.ToLower(strings.TrimSpace(os.Getenv("GOFRAME_COLOR"))); raw != "" {
		if c, ok := parseColorMode(raw); ok {
			opts.Color = c
		}
	}
	if raw := strings.ToLower(strings.TrimSpace(os.Getenv("GOFRAME_SYMBOLS"))); raw != "" {
		switch raw {
		case "1", "true", "yes", "on":
			opts.Symbols = true
		case "0", "false", "no", "off":
			opts.Symbols = false
		}
	}
	return opts
}

func currentOutputOptions() outputOptions {
	outputOptionsMu.RLock()
	defer outputOptionsMu.RUnlock()
	return outputOptionsV
}

func setOutputOptions(opts outputOptions) {
	outputOptionsMu.Lock()
	outputOptionsV = opts
	outputOptionsMu.Unlock()
}

func parseGlobalOutputOptions(args []string) ([]string, outputOptions, error) {
	opts := defaultOutputOptions()
	if len(args) == 0 {
		return args, opts, nil
	}

	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		if arg == "" {
			continue
		}
		if !strings.HasPrefix(arg, "-") {
			return args[i:], opts, nil
		}
		if arg == "--" {
			if i+1 >= len(args) {
				return nil, opts, fmt.Errorf("missing command after --")
			}
			return args[i+1:], opts, nil
		}

		if arg == "--json" {
			opts.Format = outputFormatJSON
			continue
		}
		if arg == "--symbols" {
			opts.Symbols = true
			continue
		}
		if arg == "--no-symbols" {
			opts.Symbols = false
			continue
		}

		if value, ok := extractFlagValue(arg, "--output"); ok {
			if value == "" {
				if i+1 >= len(args) {
					return nil, opts, fmt.Errorf("--output requires a value (plain|pretty|json)")
				}
				i++
				value = args[i]
			}
			parsed, valid := parseOutputFormat(strings.ToLower(strings.TrimSpace(value)))
			if !valid {
				return nil, opts, fmt.Errorf("invalid --output value %q (expected plain|pretty|json)", value)
			}
			opts.Format = parsed
			continue
		}

		if value, ok := extractFlagValue(arg, "--color"); ok {
			if value == "" {
				if i+1 >= len(args) {
					return nil, opts, fmt.Errorf("--color requires a value (auto|always|never)")
				}
				i++
				value = args[i]
			}
			parsed, valid := parseColorMode(strings.ToLower(strings.TrimSpace(value)))
			if !valid {
				return nil, opts, fmt.Errorf("invalid --color value %q (expected auto|always|never)", value)
			}
			opts.Color = parsed
			continue
		}

		// Unknown prefix flag: leave it to command-level parsers.
		return args[i:], opts, nil
	}

	return nil, opts, nil
}

func parseOutputFormat(value string) (outputFormat, bool) {
	switch outputFormat(value) {
	case outputFormatPlain:
		return outputFormatPlain, true
	case outputFormatPretty:
		return outputFormatPretty, true
	case outputFormatJSON:
		return outputFormatJSON, true
	default:
		return "", false
	}
}

func parseColorMode(value string) (colorMode, bool) {
	switch colorMode(value) {
	case colorModeAuto:
		return colorModeAuto, true
	case colorModeAlways:
		return colorModeAlways, true
	case colorModeNever:
		return colorModeNever, true
	default:
		return "", false
	}
}

func extractFlagValue(arg, name string) (string, bool) {
	if arg == name {
		return "", true
	}
	prefix := name + "="
	if strings.HasPrefix(arg, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(arg, prefix)), true
	}
	return "", false
}

func outputWantsJSON(localJSON bool) bool {
	if localJSON {
		return true
	}
	return currentOutputOptions().Format == outputFormatJSON
}

func outputIsPretty() bool {
	return currentOutputOptions().Format == outputFormatPretty
}

func useColorForWriter(w io.Writer) bool {
	opts := currentOutputOptions()
	switch opts.Color {
	case colorModeAlways:
		return true
	case colorModeNever:
		return false
	case colorModeAuto:
		if strings.EqualFold(strings.TrimSpace(os.Getenv("NO_COLOR")), "1") || os.Getenv("NO_COLOR") != "" {
			return false
		}
		if strings.EqualFold(strings.TrimSpace(os.Getenv("TERM")), "dumb") {
			return false
		}
		file, ok := w.(*os.File)
		if !ok {
			return false
		}
		info, err := file.Stat()
		if err != nil {
			return false
		}
		return (info.Mode() & os.ModeCharDevice) != 0
	default:
		return false
	}
}

func colorizeText(w io.Writer, colorCode string, text string) string {
	if !useColorForWriter(w) || strings.TrimSpace(colorCode) == "" {
		return text
	}
	return "\x1b[" + colorCode + "m" + text + "\x1b[0m"
}

func statusTag(w io.Writer, status string) string {
	normalized := strings.ToLower(strings.TrimSpace(status))
	if normalized == "" {
		normalized = "info"
	}

	if !outputIsPretty() {
		return strings.ToUpper(normalized)
	}

	opts := currentOutputOptions()
	prefix := ""
	if opts.Symbols {
		switch normalized {
		case "ok":
			prefix = "+"
		case "warning", "warn":
			prefix = "!"
		case "error", "degraded":
			prefix = "x"
		default:
			prefix = "i"
		}
	}

	label := strings.ToUpper(normalized)
	if prefix != "" {
		label = prefix + " " + label
	}

	switch normalized {
	case "ok":
		return colorizeText(w, "32", label)
	case "warning", "warn":
		return colorizeText(w, "33", label)
	case "error", "degraded":
		return colorizeText(w, "31", label)
	default:
		return colorizeText(w, "36", label)
	}
}
