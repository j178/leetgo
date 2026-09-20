package lang

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	// "github.com/pelletier/go-toml/v2"
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

	// Then: no unused src/main.rs is created.
	mainPath := filepath.Join(outDir, "src", "main.rs")
	if _, err := os.Stat(mainPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no src/main.rs, got error: %v", err)
	}
}
