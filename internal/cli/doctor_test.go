package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunDoctorReportsWarningsInsteadOfPlaceholderPasses(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	if err := runDoctor([]string{"--check", "tasks", "--json"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("doctor should allow warning-only reports: %v", err)
	}

	var report doctorReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("decode doctor report: %v", err)
	}
	if report.OverallStatus != "degraded" {
		t.Fatalf("expected degraded report, got %q", report.OverallStatus)
	}
	if report.Warnings != 1 || report.Passed != 0 {
		t.Fatalf("expected one warning and no placeholder pass, got passed=%d warnings=%d", report.Passed, report.Warnings)
	}
	if len(report.Results) != 1 || report.Results[0].Status != "warning" {
		t.Fatalf("expected warning result, got %#v", report.Results)
	}
	if strings.Contains(strings.ToLower(report.Results[0].Message), "placeholder") {
		t.Fatalf("doctor message must be honest, got %q", report.Results[0].Message)
	}
}

func TestRunDoctorRejectsUnknownCheck(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	err := runDoctor([]string{"--check", "missing"}, strings.NewReader(""), &out, &errOut)
	if err == nil {
		t.Fatal("expected unknown check error")
	}
}
