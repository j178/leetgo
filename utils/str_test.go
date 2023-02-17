package utils_test

import (
	"testing"

	"github.com/j178/leetgo/utils"
)

func TestCondenseEmptyLines(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No empty lines",
			input:    "line 1\nline 2\nline 3",
			expected: "line 1\nline 2\nline 3",
		},
		{
			name:     "One empty line",
			input:    "line 1\n\nline 2\nline 3",
			expected: "line 1\n\nline 2\nline 3",
		},
		{
			name:     "Two empty lines",
			input:    "line 1\n\n\nline 2\nline 3",
			expected: "line 1\n\nline 2\nline 3",
		},
		{
			name:     "Multiple empty lines",
			input:    "line 1\n\n\n\n\n\n\nline 2\nline 3\n\n\n\n\nline 4",
			expected: "line 1\n\nline 2\nline 3\n\nline 4",
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.name, func(t *testing.T) {
				actual := utils.CondenseEmptyLines(tc.input)
				if actual != tc.expected {
					t.Errorf("Expected result '%s', but got '%s'", tc.expected, actual)
				}
			},
		)
	}
}

func TestDecodeRawUnicodeEscape(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{
			input:  "Hello\\u0020world",
			output: "Hello world",
		},
		{
			input:  "\\u00a9 2023",
			output: "© 2023",
		},
		{
			input:  "\\u4e16\\u754c\\u60a8\\u597d",
			output: "世界您好",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.input, func(t *testing.T) {
				got := utils.DecodeRawUnicodeEscape(tt.input)
				if got != tt.output {
					t.Errorf("DecodeRawUnicodeEscape(%q) = %q; want %q", tt.input, got, tt.output)
				}
			},
		)
	}
}
