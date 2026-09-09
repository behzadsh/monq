package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// examplePackage is a real package in this module rather than a fixture: it is compiled by every build, its tests
// use the generated paths against the monq API, and the file monqgen wrote is committed alongside the struct.
const examplePackage = "./internal/example"

// TestGeneratedFileIsUpToDate regenerates the example package's paths and compares them with the committed file.
//
// It is the check the golden tests cannot make. Those compare text the generator produced against text it produced
// earlier, so they freeze whatever it emitted, working or not. This file is compiled by go build and exercised by
// the example package's own tests, so a mismatch here means the committed code and the generator have parted ways
// and one of them is wrong.
func TestGeneratedFileIsUpToDate(t *testing.T) {
	pkg, err := load(examplePackage)
	if err != nil {
		t.Fatalf("loading %s: %v", examplePackage, err)
	}

	for _, typeName := range []string{"User", "sessionDoc"} {
		t.Run(
			typeName, func(t *testing.T) {
				model, err := Parse(pkg, typeName)
				if err != nil {
					t.Fatalf("Parse() error = %v", err)
				}

				got, err := Generate(model, pkg.Name, "")
				if err != nil {
					t.Fatalf("Generate() error = %v", err)
				}

				path := filepath.Join("internal", "example", outputName(typeName, ""))

				want, err := os.ReadFile(path) //nolint:gosec // a fixed path inside the package being tested
				if err != nil {
					t.Fatalf("reading %s: %v", path, err)
				}

				if string(got) != string(want) {
					t.Errorf("%s is out of date; run go generate ./... to rewrite it", path)
				}
			},
		)
	}
}

// TestRunWritesAndRefusesToClobber drives the command the way go:generate does, then checks that a second run over
// a file monqgen did not write stops instead of replacing it.
//
// The struct it generates from uses only built-in types, so the temporary package needs nothing but a go.mod to be
// loadable. Whether the generated file compiles is settled by the example package, which the build compiles for
// real.
func TestRunWritesAndRefusesToClobber(t *testing.T) {
	dir := t.TempDir()

	write := func(name, contents string) {
		t.Helper()

		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}

	write("go.mod", "module example\n\ngo 1.25\n")
	write("user.go", tempStruct)

	if err := run(dir, "User", "", "", false); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	written, err := os.ReadFile(filepath.Join(dir, "user_paths.go")) //nolint:gosec // a fixed path in a temp dir
	if err != nil {
		t.Fatalf("reading the generated file: %v", err)
	}

	if !strings.HasPrefix(string(written), header) {
		t.Errorf("the generated file does not start with %q", header)
	}

	// Element paths are composed at run time from the prefix an array's positional methods pass in, which is why
	// "items.sku" is not in the source: the constructor holds ".sku" and the value supplies "items".
	for _, want := range []string{`"email"`, `newUserItemsPaths("items")`, `prefix + ".sku"`} {
		if !strings.Contains(string(written), want) {
			t.Errorf("the generated file is missing %s:\n%s", want, written)
		}
	}

	// The generated file imports monq, which this bare module cannot resolve, so it goes away before the package is
	// loaded again. Whether monqgen leaves its own output alone is covered by TestGuardOverwrite.
	if err := os.Remove(filepath.Join(dir, "user_paths.go")); err != nil {
		t.Fatalf("removing the generated file: %v", err)
	}

	// A file monqgen did not write is left alone.
	write("handwritten.go", "package example\n")

	if err := run(dir, "User", "", "handwritten.go", false); err == nil {
		t.Error("run() overwrote a file it did not generate, want a refusal")
	}
}

// TestRunNamesTheValue drives -var the way a go:generate line would, since the flag reaching the generated file is
// the part the generator's own tests cannot show.
func TestRunNamesTheValue(t *testing.T) {
	dir := t.TempDir()

	for name, contents := range map[string]string{"go.mod": "module example\n\ngo 1.25\n", "user.go": tempStruct} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}

	if err := run(dir, "User", "paths", "", false); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	written, err := os.ReadFile(filepath.Join(dir, "user_paths.go")) //nolint:gosec // a fixed path in a temp dir
	if err != nil {
		t.Fatalf("reading the generated file: %v", err)
	}

	if !strings.Contains(string(written), "var paths = _UserPaths{") {
		t.Errorf("the generated file does not name the value paths:\n%s", written)
	}
}

// TestRunPrints walks the tree without writing anything, which is what -print is for.
func TestRunPrints(t *testing.T) {
	if err := run(examplePackage, "User", "", "", true); err != nil {
		t.Fatalf("run() error = %v", err)
	}
}

// TestRunReportsAMissingType checks the error a mistyped -type produces.
func TestRunReportsAMissingType(t *testing.T) {
	if err := run(examplePackage, "Missing", "", "", true); err == nil {
		t.Error("run() error = nil, want an error naming the missing type")
	}
}

// tempStruct is the input the temporary package holds: built-in types only, so loading it needs no dependencies.
const tempStruct = `package example

type Item struct {
	SKU string ` + "`bson:\"sku\"`" + `
}

type User struct {
	Email string ` + "`bson:\"email\"`" + `
	Items []Item ` + "`bson:\"items\"`" + `
}
`
