package lang

import (
	"strings"
	"testing"

	"github.com/j178/leetgo/leetcode"
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

func TestRustMutableReferenceArg(t *testing.T) {
	q := &leetcode.QuestionData{
		MetaData: leetcode.MetaData{
			Name: "removeElement",
			Params: []leetcode.MetaDataParam{
				{Name: "nums", Type: "integer[]"},
				{Name: "val", Type: "integer"},
			},
			Return: &leetcode.MetaDataReturn{Type: "integer"},
		},
		CodeSnippets: []leetcode.CodeSnippet{{
			LangSlug: "rust",
			Code:     "impl Solution {pub fn remove_element(nums: &mut Vec<i32>, val:i32) -> i32 {0}}",
		}},
	}
	got, err := (rust{}).generateNormalTestCode(q)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"let mut nums: Vec<i32>",
		"Solution::remove_element(&mut nums, val)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("generated code does not contain %q", want)
		}
	}
}
