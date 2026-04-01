package cli

import (
	"testing"
	"time"

	"github.com/jcsvwinston/GoFrame/pkg/app"
)

func TestDiffConfig(t *testing.T) {
	defaultCfg := app.DefaultConfig()
	current := defaultCfg
	current.Port = 9090
	current.LogFormat = "text"
	current.ReadTimeout = 10 * time.Second

	diffs := diffConfig(defaultCfg, current)

	var changed map[string]settingDiff
	changed = make(map[string]settingDiff)
	for _, item := range diffs {
		if item.Changed {
			changed[item.Key] = item
		}
	}

	if _, ok := changed["port"]; !ok {
		t.Fatalf("expected changed setting for port")
	}
	if _, ok := changed["log_format"]; !ok {
		t.Fatalf("expected changed setting for log_format")
	}
	if _, ok := changed["read_timeout"]; !ok {
		t.Fatalf("expected changed setting for read_timeout")
	}

	if len(changed) < 3 {
		t.Fatalf("expected at least 3 changed settings, got %d", len(changed))
	}
}

func TestFormatSettingValue(t *testing.T) {
	if got := formatSettingValue(""); got != `""` {
		t.Fatalf("expected empty string formatting, got %q", got)
	}
	if got := formatSettingValue(true); got != "true" {
		t.Fatalf("expected bool formatting, got %q", got)
	}
}
