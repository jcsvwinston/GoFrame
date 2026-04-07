package contracts

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/jcsvwinston/GoFrame/internal/cli"
	"github.com/jcsvwinston/GoFrame/pkg/app"
)

func TestContractFreeze_CLIPrimaryCommands_NoRemovals(t *testing.T) {
	baseline := readBaselineLines(t, "baseline", "cli_primary_commands.txt")
	current := toSet(cli.ContractPrimaryCommandNames())

	missing := make([]string, 0)
	for _, command := range baseline {
		if _, ok := current[command]; !ok {
			missing = append(missing, command)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("stable CLI contract regression: missing primary command(s): %s", strings.Join(missing, ", "))
	}
}

func TestContractFreeze_ConfigKeyPatterns_NoRemovals(t *testing.T) {
	baseline := readBaselineLines(t, "baseline", "config_key_patterns.txt")
	current := toSet(app.ContractConfigKeyPatterns())

	missing := make([]string, 0)
	for _, key := range baseline {
		if _, ok := current[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("stable config contract regression: missing key pattern(s): %s", strings.Join(missing, ", "))
	}
}

func TestContractFreeze_BaselinesAreSortedUnique(t *testing.T) {
	checkSortedUnique(t, readBaselineLines(t, "baseline", "cli_primary_commands.txt"), "cli_primary_commands.txt")
	checkSortedUnique(t, readBaselineLines(t, "baseline", "config_key_patterns.txt"), "config_key_patterns.txt")
}

func checkSortedUnique(t *testing.T, lines []string, name string) {
	t.Helper()
	seen := map[string]struct{}{}
	for i, line := range lines {
		if i > 0 && lines[i-1] > line {
			t.Fatalf("%s must be sorted ascending; %q appears before %q", name, lines[i-1], line)
		}
		if _, exists := seen[line]; exists {
			t.Fatalf("%s contains duplicate entry %q", name, line)
		}
		seen[line] = struct{}{}
	}
}

func readBaselineLines(t *testing.T, rel ...string) []string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve current file path")
	}
	base := filepath.Join(filepath.Dir(file), filepath.Join(rel...))
	f, err := os.Open(base)
	if err != nil {
		t.Fatalf("open baseline %s: %v", base, err)
	}
	defer f.Close()

	out := make([]string, 0, 64)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan baseline %s: %v", base, err)
	}
	return out
}

func toSet(items []string) map[string]struct{} {
	out := make(map[string]struct{}, len(items))
	for _, item := range items {
		out[item] = struct{}{}
	}
	return out
}
