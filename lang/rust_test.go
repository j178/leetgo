package lang

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

func TestToRustVarName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"query", "query"},
		{"isValidBST", "is_valid_bst"},
	}
	for _, tt := range tests {
		if got := toRustVarName(tt.name); got != tt.want {
			t.Errorf("toRustVarName(%v) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestRustInitWorkspace(t *testing.T) {
	// Given: new isolated home directory.
	homeDir := t.TempDir()
	t.Setenv("LEETGO_HOME", homeDir)

	if err := os.Mkdir(filepath.Join(homeDir, "cache"), 0o700); err != nil {
		t.Fatal(err)
	}

	// When: new Rust workspace is initialized.
	outDir := t.TempDir()
	if err := rustGen.InitWorkspace(outDir); err != nil {
		t.Fatalf("initialize Rust workspace: %v", err)
	}

	// Then: Cargo.toml contains the expected package and dependencies.
	tomlPath := filepath.Join(outDir, "Cargo.toml")
	data, err := os.ReadFile(tomlPath)
	if err != nil {
		t.Fatalf("read Cargo.toml: %v", err)
	}

	var manifest struct {
		Package struct {
			Name    string `toml:"name"`
			Version string `toml:"version"`
			Edition string `toml:"edition"`
		} `toml:"package"`
		Dependencies map[string]string `toml:"dependencies"`
	}
	if err := toml.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("parse Cargo.toml: %v", err)
	}

	if manifest.Package.Name != "leetcode-solutions" {
		t.Errorf("unexpected package name: %q", manifest.Package.Name)
	}
	if manifest.Package.Version != "0.1.0" {
		t.Errorf("unexpected package version: %q", manifest.Package.Version)
	}
	if manifest.Package.Edition != "2024" {
		t.Errorf("unexpected Rust package Rust edition: %q", manifest.Package.Edition)
	}

	expectedDeps := map[string]string{
		"serde":      "1.0.196",
		"serde_json": "1.0.113",
		"anyhow":     "1.0.79",
	}
	for depName, expectedDepVersion := range expectedDeps {
		if DepVersion := manifest.Dependencies[depName]; DepVersion != expectedDepVersion {
			t.Errorf("dependency: %s: got: %q, want: %q", depName, DepVersion, expectedDepVersion)
		}
	}

	// And: no unused src/main.rs is created.
	mainPath := filepath.Join(outDir, "src", "main.rs")
	if _, err := os.Stat(mainPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no src/main.rs, got error: %v", err)
	}
}
