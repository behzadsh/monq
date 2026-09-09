package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// update rewrites the golden files instead of comparing against them: go test ./cmd/monqgen -update.
var update = flag.Bool("update", false, "rewrite the golden files")

// goldenPath is where the expected output of one type lives. The extension keeps it out of the build, since the
// file is Go source for another package.
func goldenPath(typeName string) string {
	return filepath.Join("testdata", "golden", strings.ToLower(typeName)+"_paths.golden")
}

func TestGenerate(t *testing.T) {
	for _, typeName := range []string{"User", "Custom", "unexportedDoc"} {
		t.Run(
			typeName, func(t *testing.T) {
				model, err := Parse(loadFixtures(t), typeName)
				if err != nil {
					t.Fatalf("Parse() error = %v", err)
				}

				got, err := Generate(model, "fixtures", "")
				if err != nil {
					t.Fatalf("Generate() error = %v", err)
				}

				golden := goldenPath(typeName)

				if *update {
					if writeErr := os.WriteFile(golden, got, 0o600); writeErr != nil {
						t.Fatalf("writing %s: %v", golden, writeErr)
					}

					return
				}

				want, err := os.ReadFile(golden) //nolint:gosec // the path is built from a fixed test directory
				if err != nil {
					t.Fatalf("reading %s: %v (run go test ./cmd/monqgen -update to create it)", golden, err)
				}

				if string(got) != string(want) {
					t.Errorf("generated source differs from %s\n--- got ---\n%s", golden, got)
				}
			},
		)
	}
}

func TestGenerateRefusesAReservedName(t *testing.T) {
	model, err := Parse(loadFixtures(t), "Reserved")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	_, err = Generate(model, "fixtures", "")
	if err == nil {
		t.Fatal("Generate() error = nil, want an error about the reserved field name")
	}

	if !strings.Contains(err.Error(), "Path") {
		t.Fatalf("Generate() error = %v, want it to name the colliding field", err)
	}
}

func TestGenerateProducesFormattedSource(t *testing.T) {
	model, err := Parse(loadFixtures(t), "User")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	got, err := Generate(model, "fixtures", "")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.HasPrefix(string(got), header) {
		t.Errorf("generated source does not start with %q", header)
	}

	// Generate runs the source through go/format, so a second pass has nothing left to change.
	second, err := Generate(model, "fixtures", "")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if string(got) != string(second) {
		t.Error("Generate() is not deterministic across two runs")
	}
}

func TestGenerateNamesTheValue(t *testing.T) {
	model, err := Parse(loadFixtures(t), "User")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	got, err := Generate(model, "fixtures", "paths")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Only the value is renamed: the types go on being named after the model, since nothing outside the file reads
	// them.
	for _, want := range []string{"var paths = _UserPaths{", "type _UserPaths struct {"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("generated source is missing %q:\n%s", want, got)
		}
	}

	if strings.Contains(string(got), "var UserPaths") {
		t.Errorf("generated source still declares the default value name:\n%s", got)
	}
}

func TestGenerateRefusesAValueNamedAfterAType(t *testing.T) {
	model, err := Parse(loadFixtures(t), "User")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	_, err = Generate(model, "fixtures", "_UserPaths")
	if err == nil {
		t.Fatal("Generate() error = nil, want an error about the name the root type has")
	}
}

func TestCheckVarName(t *testing.T) {
	tests := []struct {
		name      string
		varName   string
		wantError bool
	}{
		{name: "empty stands for the default"},
		{name: "an identifier", varName: "campaignDocPaths"},
		{name: "an underscore is allowed", varName: "_paths"},
		{name: "a space is not", varName: "campaign paths", wantError: true},
		{name: "a dot is not", varName: "brand.Paths", wantError: true},
		{name: "a keyword is not", varName: "type", wantError: true},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := checkVarName(tt.varName)
				if tt.wantError && err == nil {
					t.Fatalf("checkVarName(%q) error = nil, want a refusal", tt.varName)
				}

				if !tt.wantError && err != nil {
					t.Fatalf("checkVarName(%q) error = %v, want nil", tt.varName, err)
				}
			},
		)
	}
}

func TestTypeNamesDoNotCollide(t *testing.T) {
	g := &generator{names: map[string]bool{}}

	tests := []struct {
		base string
		want string
	}{
		{base: "User", want: "_UserPaths"},
		{base: "User", want: "_UserPaths2"},
		{base: "User", want: "_UserPaths3"},
		{base: "Order", want: "_OrderPaths"},
	}

	for _, tt := range tests {
		t.Run(
			tt.want, func(t *testing.T) {
				if got := g.typeName(tt.base); got != tt.want {
					t.Fatalf("typeName(%q) = %q, want %q", tt.base, got, tt.want)
				}
			},
		)
	}
}

func TestGuardOverwrite(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name      string
		file      string
		contents  string
		wantError bool
	}{
		{name: "a file that is not there yet", file: "new.go"},
		{name: "a file monqgen wrote", file: "generated.go", contents: header + "\n\npackage x\n"},
		{name: "a file someone else wrote", file: "handwritten.go", contents: "package x\n", wantError: true},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				path := filepath.Join(dir, tt.file)

				if tt.contents != "" {
					if err := os.WriteFile(path, []byte(tt.contents), 0o600); err != nil {
						t.Fatalf("writing %s: %v", path, err)
					}
				}

				err := guardOverwrite(path)
				if tt.wantError && err == nil {
					t.Fatal("guardOverwrite() error = nil, want a refusal")
				}

				if !tt.wantError && err != nil {
					t.Fatalf("guardOverwrite() error = %v, want nil", err)
				}
			},
		)
	}
}

func TestOutputName(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		out      string
		want     string
	}{
		{name: "default is the type lowercased", typeName: "User", want: "user_paths.go"},
		{name: "one file per type", typeName: "OrderLine", want: "orderline_paths.go"},
		{name: "an explicit name wins", typeName: "User", out: "paths.go", want: "paths.go"},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := outputName(tt.typeName, tt.out); got != tt.want {
					t.Fatalf("outputName(%q, %q) = %q, want %q", tt.typeName, tt.out, got, tt.want)
				}
			},
		)
	}
}
