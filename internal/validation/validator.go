// internal/validation/validator.go
package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

const SchemaRelPath = "schemas/observatory_event.schema.json"

type Validator struct {
	schema *jsonschema.Schema
}

func NewValidator(repoRoot string) (*Validator, error) {
	// Compile from a local file path (no go:embed, no .. patterns)
	schemaPath := filepath.Join(repoRoot, SchemaRelPath)
	compiler := jsonschema.NewCompiler()
	// Load from file with a file:// URL to satisfy the compiler’s resolver
	u := "file://" + filepath.ToSlash(schemaPath)
	if err := compiler.AddResource(u, mustOpen(schemaPath)); err != nil {
		return nil, fmt.Errorf("add resource: %w", err)
	}
	s, err := compiler.Compile(u)
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}
	return &Validator{schema: s}, nil
}

func (v *Validator) ValidateBytes(b []byte) error {
	var any map[string]any
	if err := json.Unmarshal(b, &any); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	if err := v.schema.Validate(any); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}
	return nil
}

func mustOpen(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	return f
}