// internal/validation/validator_test.go
package validation

import (
	"os"
	"path/filepath"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	// Walk up from this file to project root (two levels up from /internal/validation)
	dir, err := os.Getwd()
	if err != nil { t.Fatal(err) }
	// If tests run from module root, keep "."
	// Otherwise, find a dir that contains "schemas/observatory_event.schema.json"
	for i := 0; i < 4; i++ {
		if _, err := os.Stat(filepath.Join(dir, SchemaRelPath)); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatalf("could not locate repo root containing %s", SchemaRelPath)
	return ""
}

func Test_ValidSample(t *testing.T) {
	root := repoRoot(t)
	v, err := NewValidator(root)
	if err != nil { t.Fatal(err) }

	path := filepath.Join(root, "sampledata", "sample_event.json")
	b, err := os.ReadFile(path)
	if err != nil { t.Fatal(err) }

	if err := v.ValidateBytes(b); err != nil {
		t.Fatalf("expected valid, got error: %v", err)
	}
}

func Test_InvalidSample(t *testing.T) {
	root := repoRoot(t)
	v, err := NewValidator(root)
	if err != nil { t.Fatal(err) }

	path := filepath.Join(root, "sampledata", "sample_event_invalid.json")
	b, err := os.ReadFile(path)
	if err != nil { t.Fatal(err) }

	if err := v.ValidateBytes(b); err == nil {
		t.Fatalf("expected invalid sample to fail validation")
	}
}