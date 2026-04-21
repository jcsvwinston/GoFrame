package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

const contractsAggregatorTemplate = `package contracts

import "github.com/jcsvwinston/GoFrame/pkg/openapi"

type Registrar func(doc *openapi.Document)

var registrars []Registrar

func RegisterContract(register Registrar) {
	if register == nil {
		return
	}
	registrars = append(registrars, register)
}

func Register(doc *openapi.Document) {
	if doc == nil {
		return
	}
	for _, register := range registrars {
		register(doc)
	}
}

func NewDocument() *openapi.Document {
	doc := openapi.NewDocument(%q, "0.1.0")
	Register(doc)
	return doc
}
`

func ensureContractsAggregator(outDir, title string) error {
	path := filepath.Join(outDir, "internal", "contracts", "contracts.go")
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat contracts aggregator: %w", err)
	}

	if title == "" {
		title = "Application API"
	}

	body := fmt.Sprintf(contractsAggregatorTemplate, title)
	return writeFileIfNotExists(path, body, false)
}

func defaultOpenAPITitle(projectName, modulePath, outDir string) string {
	if projectName = filepath.Base(projectName); projectName != "" && projectName != "." {
		return projectName + " API"
	}
	if modulePath != "" {
		return toPascalCase(filepath.Base(modulePath)) + " API"
	}
	if outDir != "" {
		base := filepath.Base(filepath.Clean(outDir))
		if base != "" && base != "." && base != string(filepath.Separator) {
			return toPascalCase(base) + " API"
		}
	}
	return "Application API"
}
