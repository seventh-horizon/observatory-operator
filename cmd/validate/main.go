// cmd/validate/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/observatory-operator/internal/validation"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <path-to-json>\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	repoRoot := flag.String("root", ".", "repo root (folder containing schemas/)")
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	target := flag.Arg(0)

	v, err := validation.NewValidator(*repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init validator: %v\n", err)
		os.Exit(1)
	}
	b, err := os.ReadFile(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", target, err)
		os.Exit(1)
	}
	if err := v.ValidateBytes(b); err != nil {
		fmt.Fprintf(os.Stderr, "INVALID: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK")
}