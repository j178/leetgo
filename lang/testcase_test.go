package lang

import (
	"testing"

	"github.com/j178/leetgo/leetcode"
)

func TestParseTestCasesWithoutQuestionMetadata(t *testing.T) {
	q := &leetcode.QuestionData{TitleSlug: "two-sum"}
	f := &FileOutput{Content: `input:
[1,2]
output:
3
`}

	cases, err := ParseTestCases(q, f)
	if err != nil {
		t.Fatalf("ParseTestCases() error = %v", err)
	}
	if len(cases.Cases) != 1 || cases.Cases[0].Output != "3" {
		t.Fatalf("unexpected test cases: %+v", cases.Cases)
	}
}
