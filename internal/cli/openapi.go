package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runOpenAPI(args []string, _ io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("openapi", flag.ContinueOnError)
	fs.SetOutput(stderr)

	outPath := fs.String("out", "openapi.json", "Output path for the exported OpenAPI JSON document, or - for stdout")
	projectDir := fs.String("project", ".", "Project root that contains go.mod and internal/contracts")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if len(fs.Args()) != 0 {
		return fmt.Errorf("usage: goframe openapi [--project .] [--out openapi.json]")
	}

	root, err := filepath.Abs(strings.TrimSpace(*projectDir))
	if err != nil {
		return fmt.Errorf("resolve project path: %w", err)
	}

	modulePath, hasModule, err := detectModulePath(root)
	if err != nil {
		return err
	}
	if !hasModule {
		return fmt.Errorf("openapi export requires a Go module in %s", root)
	}

	exporterDir, err := os.MkdirTemp(root, ".goframe-openapi-*")
	if err != nil {
		return fmt.Errorf("create exporter workspace: %w", err)
	}
	defer os.RemoveAll(exporterDir)

	exporterMainPath := filepath.Join(exporterDir, "main.go")
	exporterMainBody := fmt.Sprintf(openAPIExporterTemplate, modulePath)
	if err := os.WriteFile(exporterMainPath, []byte(exporterMainBody), 0644); err != nil {
		return fmt.Errorf("write exporter entrypoint: %w", err)
	}

	exportRel, err := filepath.Rel(root, exporterDir)
	if err != nil {
		return fmt.Errorf("resolve exporter path: %w", err)
	}

	cmd := exec.Command("go", "run", "./"+filepath.ToSlash(exportRel))
	cmd.Dir = root

	var body bytes.Buffer
	var cmdErr bytes.Buffer
	cmd.Stdout = &body
	cmd.Stderr = &cmdErr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(cmdErr.String())
		if msg == "" {
			return fmt.Errorf("export openapi document: %w", err)
		}
		return fmt.Errorf("export openapi document: %w: %s", err, msg)
	}
	if !json.Valid(body.Bytes()) {
		return fmt.Errorf("openapi export produced invalid JSON")
	}

	if strings.TrimSpace(*outPath) == "-" {
		if _, err := stdout.Write(body.Bytes()); err != nil {
			return fmt.Errorf("write openapi stdout: %w", err)
		}
		if body.Len() > 0 && body.Bytes()[body.Len()-1] != '\n' {
			_, _ = io.WriteString(stdout, "\n")
		}
		return nil
	}

	targetPath := strings.TrimSpace(*outPath)
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(root, targetPath)
	}
	if err := ensureDir(filepath.Dir(targetPath)); err != nil {
		return err
	}
	if err := os.WriteFile(targetPath, body.Bytes(), 0644); err != nil {
		return fmt.Errorf("write openapi document %s: %w", targetPath, err)
	}

	fmt.Fprintf(stdout, "OpenAPI document exported: %s\n", targetPath)
	return nil
}

const openAPIExporterTemplate = `package main

import (
	"log"
	"os"

	"%[1]s/internal/contracts"
	"github.com/jcsvwinston/nucleus/pkg/openapi"
)

func main() {
	doc := contracts.NewDocument()
	if err := openapi.WriteJSON(os.Stdout, doc); err != nil {
		log.Fatal(err)
	}
}
`
