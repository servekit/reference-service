// gen compiles the reference tables in internal/data from pinned upstreams:
// CLDR 45.0.0 (unpkg split packages), the ISO 3166 dataset, and IANA tzdb.
// Needs network on first run; results are cached under .cache (gitignored).
// Data updates flow: run `make data`, review the diff, PR, release.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var outDir = flag.String("out", "internal/data", "target directory for generated *_data.go files")

// outputPath resolves a generated file name against -out.
func outputPath(name string) string { return filepath.Join(*outDir, name) }

func main() {
	flag.Parse()
	cache := filepath.Join("tools", "gen", ".cache")

	if err := checkDataLocales(*outDir); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}

	steps := []struct {
		name string
		fn   func(string) error
	}{
		{"countries", buildCountries},
		{"timezones", buildTimezones},
		{"languages", buildLanguages},
		{"currencies", buildCurrencies},
		{"region_groups", buildRegionGroups},
	}
	for _, s := range steps {
		if err := s.fn(cache); err != nil {
			fmt.Fprintf(os.Stderr, "gen %s: %v\n", s.name, err)
			os.Exit(1)
		}
		fmt.Printf("gen: %s ok\n", s.name)
	}
}
