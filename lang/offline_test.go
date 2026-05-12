package lang

import (
	"testing"

	"github.com/j178/leetgo/config"
)

func TestParseOfflineTestCases(t *testing.T) {
	f := &FileOutput{
		Content: `input:
[1,2]
3
output:
5

input:
[4]
7
output:
11
`,
	}

	tc, err := ParseOfflineTestCases(f)
	if err != nil {
		t.Fatalf("ParseOfflineTestCases() error = %v", err)
	}
	if got, want := len(tc.Cases), 2; got != want {
		t.Fatalf("ParseOfflineTestCases() cases = %d, want %d", got, want)
	}
	if got, want := tc.Cases[0].Output, "5"; got != want {
		t.Fatalf("first case output = %q, want %q", got, want)
	}
	if got, want := tc.Cases[1].InputString(), "[4]\n7\n"; got != want {
		t.Fatalf("second case input = %q, want %q", got, want)
	}
}

func TestNewOfflineGenerateResult(t *testing.T) {
	entry := config.OfflineQuestion{
		FrontendID:    "1",
		Slug:          "two-sum",
		Lang:          "go",
		OutDir:        "/tmp/leetgo/go",
		SubDir:        "0001.two-sum",
		CodeFile:      "solution.go",
		TestCasesFile: "testcases.txt",
	}

	result := NewOfflineGenerateResult(entry, golang{})
	if got, want := result.GetFile(TestFile).GetPath(), "/tmp/leetgo/go/0001.two-sum/solution.go"; got != want {
		t.Fatalf("code path = %q, want %q", got, want)
	}
	if got, want := result.GetFile(TestCasesFile).GetPath(), "/tmp/leetgo/go/0001.two-sum/testcases.txt"; got != want {
		t.Fatalf("testcases path = %q, want %q", got, want)
	}
}
