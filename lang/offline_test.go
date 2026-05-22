package lang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/j178/leetgo/config"
)

func prepareOfflineResolveTest(t *testing.T) {
	t.Helper()

	root := t.TempDir()
	home := t.TempDir()
	t.Setenv("LEETGO_HOME", home)

	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldwd)
	})
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "leetgo.yaml"), []byte(""), 0o644); err != nil {
		t.Fatalf("write config error = %v", err)
	}
}

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

func TestParseOfflineTestCasesMissingOutput(t *testing.T) {
	f := &FileOutput{
		Content: `input:
[1,2]
3
output:

`,
	}

	if _, err := ParseOfflineTestCases(f); err == nil {
		t.Fatalf("ParseOfflineTestCases() error = nil, want failure for missing output")
	}
}

func TestResolveOfflineQuestionExactLang(t *testing.T) {
	prepareOfflineResolveTest(t)

	goEntry := config.OfflineQuestion{
		FrontendID: "1",
		Slug:       "two-sum",
		Lang:       "go",
	}
	cppEntry := config.OfflineQuestion{
		FrontendID: "1",
		Slug:       "two-sum",
		Lang:       "cpp",
	}
	if err := config.SaveOfflineQuestion(goEntry); err != nil {
		t.Fatalf("SaveOfflineQuestion(go) error = %v", err)
	}
	if err := config.SaveOfflineQuestion(cppEntry); err != nil {
		t.Fatalf("SaveOfflineQuestion(cpp) error = %v", err)
	}

	gotGo, err := config.ResolveOfflineQuestion("two-sum", "go")
	if err != nil {
		t.Fatalf("ResolveOfflineQuestion(go) error = %v", err)
	}
	if got, want := gotGo.Lang, "go"; got != want {
		t.Fatalf("ResolveOfflineQuestion(go).Lang = %q, want %q", got, want)
	}

	gotCpp, err := config.ResolveOfflineQuestion("two-sum", "cpp")
	if err != nil {
		t.Fatalf("ResolveOfflineQuestion(cpp) error = %v", err)
	}
	if got, want := gotCpp.Lang, "cpp"; got != want {
		t.Fatalf("ResolveOfflineQuestion(cpp).Lang = %q, want %q", got, want)
	}
}

func TestResolveOfflineQuestionDirectQIDIgnoresLastAndLang(t *testing.T) {
	prepareOfflineResolveTest(t)

	twoSum := config.OfflineQuestion{
		FrontendID: "2",
		Slug:       "two-sum",
		Lang:       "go",
	}
	threeSum := config.OfflineQuestion{
		FrontendID: "3",
		Slug:       "three-sum",
		Lang:       "cpp",
	}
	if err := config.SaveOfflineQuestion(twoSum); err != nil {
		t.Fatalf("SaveOfflineQuestion(twoSum) error = %v", err)
	}
	if err := config.SaveOfflineQuestion(threeSum); err != nil {
		t.Fatalf("SaveOfflineQuestion(threeSum) error = %v", err)
	}

	config.SaveState(config.State{
		LastQuestion: config.LastQuestion{
			FrontendID: "3",
			Slug:       "three-sum",
			Gen:        "cpp",
		},
	})

	gotTwo, err := config.ResolveOfflineQuestion("2", "python")
	if err != nil {
		t.Fatalf("ResolveOfflineQuestion(2) error = %v", err)
	}
	if got, want := gotTwo.Lang, "go"; got != want {
		t.Fatalf("ResolveOfflineQuestion(2).Lang = %q, want %q", got, want)
	}

	gotThree, err := config.ResolveOfflineQuestion("3", "python")
	if err != nil {
		t.Fatalf("ResolveOfflineQuestion(3) error = %v", err)
	}
	if got, want := gotThree.Lang, "cpp"; got != want {
		t.Fatalf("ResolveOfflineQuestion(3).Lang = %q, want %q", got, want)
	}

	gotLast, err := config.ResolveOfflineQuestion("last", "python")
	if err != nil {
		t.Fatalf("ResolveOfflineQuestion(last) error = %v", err)
	}
	if got, want := gotLast.Lang, "cpp"; got != want {
		t.Fatalf("ResolveOfflineQuestion(last).Lang = %q, want %q", got, want)
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
