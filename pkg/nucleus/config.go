// Package nucleus — config.go implements the configuration loader
// surfaced by `AppBuilder.FromConfigFile`. ADR-010 §2 names this as
// Phase 2 work. This file lands Phase 2a (single-file load with
// size cap, schema strict-unknown-fields, and did-you-mean hints
// for typos); multi-file merge with the `_append`/`_remove` suffix
// operators ships in Phase 2b. The package-level `Run(App)` and the
// direct-struct surface never traverse this loader — only the
// builder-chain `FromConfigFile` does.
package nucleus

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jcsvwinston/nucleus/pkg/app"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

// MaxConfigFileBytes is the per-file size cap enforced by
// FromConfigFile before invoking the YAML parser. The cap is the
// ADR-010 §17 compliance item — it eliminates the parser-DoS class
// (anchor expansion / deep nesting) that `gopkg.in/yaml.v3` is not
// hardened against by itself. 1 MiB is generous for application
// configuration in practice while still small enough to make a
// pathological file fail loud rather than wedge the process.
const MaxConfigFileBytes = 1 << 20 // 1 MiB

// ErrConfigFileTooLarge is returned when a configuration file exceeds
// MaxConfigFileBytes. Callers can errors.Is against this sentinel to
// distinguish a configuration-management problem (file is genuinely
// too big — split it) from a parser-side problem (bad YAML).
var ErrConfigFileTooLarge = errors.New("nucleus: configuration file exceeds the per-file size cap")

// ErrUnsupportedConfigFormat is returned when FromConfigFile is asked
// to parse a file whose extension is not yet supported. Phase 2a
// supports .yaml / .yml only; .toml and .json land in a Phase 2b
// follow-up that adds the corresponding parsers and updates this
// sentinel's call sites.
var ErrUnsupportedConfigFormat = errors.New("nucleus: unsupported configuration file format")

// ErrUnknownConfigKeys is returned when strict schema validation
// (the default for FromConfigFile) finds keys in the loaded file
// that do not map to any field on `app.Config` or its nested
// structs. The error's Error() reproduces the offending keys with
// "did you mean …?" hints when a close match exists.
var ErrUnknownConfigKeys = errors.New("nucleus: unknown configuration key(s)")

// loadFromFile is the Phase 2a single-file loader used by
// AppBuilder.FromConfigFile. It returns an *app.Config populated from
// the file plus the framework's struct defaults — the same precedence
// chain LoadConfig in pkg/app uses, minus the environment-variable
// layer (env vars stay attached to LoadConfig's path; FromConfigFile
// is exclusively about files for now).
//
// Validation:
//
//   - Layer 1 (syntactic): YAML parse with a 1 MiB file size cap
//     enforced before reading.
//   - Layer 2 (schema, strict): every yaml key must map to a
//     `koanf:"..."`-tagged field on `app.Config`. Unknown keys
//     produce ErrUnknownConfigKeys with did-you-mean hints.
//
// Layers 3–5 (semantic, referential, module-specific) land in later
// Phase 2 sub-iterations.
func loadFromFile(path string) (*app.Config, error) {
	if path == "" {
		return nil, errors.New("nucleus: FromConfigFile path is empty")
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		// supported
	case ".toml", ".json":
		return nil, fmt.Errorf("%w: %s parsing is a Phase 2b deliverable (path=%q)", ErrUnsupportedConfigFormat, ext, path)
	default:
		return nil, fmt.Errorf("%w: extension %q (path=%q)", ErrUnsupportedConfigFormat, ext, path)
	}

	data, err := readFileWithCap(path, MaxConfigFileBytes)
	if err != nil {
		return nil, err
	}

	// Phase 2a uses two koanf instances: one for the file content alone
	// (to enumerate keys for the strict-schema check) and one for the
	// final layered load (defaults < file). Layering through a single
	// koanf would merge before the strict check could fire — we'd lose
	// the visibility into which keys came from the user.
	fileK := koanf.New(".")
	if err := fileK.Load(rawbytes.Provider(data), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("nucleus: parse %s: %w", path, err)
	}

	// Layer 2: strict schema. Anything in the file that does not appear
	// in the framework's schema-key set is a typo or stale config; fail
	// loud with did-you-mean hints.
	schemaKeys := app.ContractConfigKeyPatterns()
	if unknown := unknownKeys(fileK.All(), schemaKeys); len(unknown) > 0 {
		return nil, formatUnknownKeys(unknown, schemaKeys)
	}

	// Combined load: struct defaults < file. Mirrors app.LoadConfig but
	// scoped to file-only (the env-var layer remains owned by
	// app.LoadConfig for callers that want it).
	k := koanf.New(".")
	if err := k.Load(structs.Provider(defaultsForConfig(), "koanf"), nil); err != nil {
		return nil, fmt.Errorf("nucleus: load defaults: %w", err)
	}
	if err := k.Load(rawbytes.Provider(data), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("nucleus: re-parse %s: %w", path, err)
	}

	var cfg app.Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("nucleus: unmarshal %s: %w", path, err)
	}
	return &cfg, nil
}

// readFileWithCap reads up to capBytes+1 bytes from path. When the
// file is larger than capBytes (the +1 is the overshoot signalling),
// it returns ErrConfigFileTooLarge wrapped with the path. Stat is not
// used as the only check because some filesystems (procfs, FUSE) lie
// about file size; reading is the source of truth.
func readFileWithCap(path string, capBytes int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("nucleus: open %s: %w", path, err)
	}
	defer f.Close()

	limited := io.LimitReader(f, capBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("nucleus: read %s: %w", path, err)
	}
	if int64(len(data)) > capBytes {
		return nil, fmt.Errorf("%w (path=%q, cap=%d bytes)", ErrConfigFileTooLarge, path, capBytes)
	}
	return data, nil
}

// defaultsForConfig returns the same defaults app.LoadConfig uses,
// reached through the public app.DefaultConfig accessor so this
// package does not need to import pkg/app's internals.
func defaultsForConfig() app.Config {
	return app.DefaultConfig()
}

// unknownKeys returns the leaf keys present in the file-koanf's
// flattened map that do NOT appear in any schemaKey prefix. The
// `app.ContractConfigKeyPatterns()` set is the canonical schema
// surface — it enumerates the koanf-bindable keys
// `pkg/app.Config` and its nested structs expose.
//
// A key matches the schema if any schemaKey is either equal to the
// key or is a prefix that the koanf flattening expanded into. Map-
// typed schema slots (like `databases.<alias>.url`) are represented
// in the patterns set as `databases.*.url`; we recognise these via a
// segment-by-segment match where `*` is a wildcard.
func unknownKeys(loaded map[string]any, schemaKeys []string) []string {
	patterns := compileKeyPatterns(schemaKeys)
	var unknown []string
	for k := range loaded {
		if !keyMatchesAny(k, patterns) {
			unknown = append(unknown, k)
		}
	}
	sort.Strings(unknown)
	return unknown
}

// compiledKeyPattern is the segment-by-segment shape used by
// keyMatchesAny. `*` segments are wildcards; everything else must
// match literally.
type compiledKeyPattern []string

func compileKeyPatterns(patterns []string) []compiledKeyPattern {
	out := make([]compiledKeyPattern, 0, len(patterns))
	for _, p := range patterns {
		out = append(out, strings.Split(p, "."))
	}
	return out
}

// keyMatchesAny reports whether key matches at least one of the
// supplied patterns. Matching is segment-by-segment with `*` as a
// single-segment wildcard.
func keyMatchesAny(key string, patterns []compiledKeyPattern) bool {
	segments := strings.Split(key, ".")
	for _, pat := range patterns {
		if len(pat) != len(segments) {
			continue
		}
		match := true
		for i, p := range pat {
			if p == "*" {
				continue
			}
			if p != segments[i] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// formatUnknownKeys produces an ErrUnknownConfigKeys-wrapped error
// listing every unknown key with a did-you-mean hint when a close
// match exists in the schema (within a Levenshtein-style edit
// distance of 3 on the deepest-segment basis).
func formatUnknownKeys(unknown, schemaKeys []string) error {
	var b strings.Builder
	for _, k := range unknown {
		b.WriteString("\n  - ")
		b.WriteString(k)
		if hint := didYouMean(k, schemaKeys); hint != "" {
			b.WriteString(" (did you mean ")
			b.WriteString(hint)
			b.WriteString("?)")
		}
	}
	// Wrap the sentinel — its Error() text already names the
	// "unknown configuration key(s)" preamble; we append the bullet
	// list rather than re-stating the preamble.
	return fmt.Errorf("%w:%s", ErrUnknownConfigKeys, b.String())
}

// didYouMean returns the closest schema key to `unknown` within an
// edit-distance threshold of 2 on the final segment, or the empty
// string when no schema key is close enough. The intent is to catch
// typos like `loging.level` → `logging.level` without producing
// noisy false-positive hints.
func didYouMean(unknown string, schemaKeys []string) string {
	uTail := lastSegment(unknown)
	if uTail == "" {
		return ""
	}
	best := ""
	bestDist := 4 // accept distance ≤3; reject 4+
	for _, k := range schemaKeys {
		sTail := lastSegment(k)
		if sTail == "" {
			continue
		}
		d := levenshtein(uTail, sTail)
		if d < bestDist {
			bestDist = d
			best = k
		}
	}
	return best
}

func lastSegment(k string) string {
	if i := strings.LastIndex(k, "."); i >= 0 {
		return k[i+1:]
	}
	return k
}

// levenshtein computes the edit distance between two ASCII strings.
// Simple O(n*m) DP — config keys are short (rarely >30 chars), so
// the allocation cost is negligible compared with the readability
// win of a textbook implementation.
func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			curr[j] = del
			if ins < curr[j] {
				curr[j] = ins
			}
			if sub < curr[j] {
				curr[j] = sub
			}
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}
