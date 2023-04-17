package lang

import "testing"

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
