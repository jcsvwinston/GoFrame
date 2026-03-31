package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"time"
)

type healthComponent struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Details string `json:"details,omitempty"`
}

type healthReport struct {
	Status     string            `json:"status"`
	CheckedAt  string            `json:"checked_at"`
	Components []healthComponent `json:"components"`
}

func runHealth(args []string, _ io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("health", flag.ContinueOnError)
	fs.SetOutput(stderr)

	configPath := fs.String("config", "", "Path to goframe config file")
	timeout := fs.Duration("timeout", 3*time.Second, "Health check timeout")
	asJSON := fs.Bool("json", false, "Print output as JSON")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if len(fs.Args()) > 0 {
		return fmt.Errorf("health does not accept positional arguments")
	}
	if *timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}

	cfg, database, cleanup, err := newDatabase(*configPath)
	if err != nil {
		return err
	}
	defer cleanup()

	report := healthReport{
		Status:    "ok",
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
		Components: []healthComponent{
			{Name: "database", Status: "ok", Details: fmt.Sprintf("engine=%s", cfg.DatabaseEngine)},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	if err := database.Health(ctx); err != nil {
		report.Status = "degraded"
		report.Components[0].Status = "error"
		report.Components[0].Details = err.Error()
	}

	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return err
		}
	} else {
		fmt.Fprintf(stdout, "overall\t%s\n", report.Status)
		for _, c := range report.Components {
			fmt.Fprintf(stdout, "%s\t%s\t%s\n", c.Name, c.Status, c.Details)
		}
	}

	if report.Status != "ok" {
		return fmt.Errorf("health check failed")
	}
	return nil
}
