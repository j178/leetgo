package lang

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	goutils "github.com/j178/leetgo/testutils/go"

	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

type TestCase struct {
	Question *leetcode.QuestionData
	No       int
	Input    []string
	Output   string
}

func (c *TestCase) Check() error {
	q := c.Question
	err := q.Fulfill()
	if err != nil {
		return fmt.Errorf("failed to get question data: %w", err)
	}
	narg := q.MetaData.NArg()
	if q.MetaData.SystemDesign {
		// System design questions have two inputs, the first one is a list of strings, but the second is a list of
		// different types. We just check if it's a valid list.
		// input:
		// ["LRUCache","put","put","get","put","get","put","get","get","get"]
		// [[2],[1,1],[2,2],[1],[3,3],[2],[4,4],[1],[3],[4]]
		// output:
		// [null,null,null,1,null,-1,null,-1,3,4]
		if len(c.Input) != narg {
			return fmt.Errorf("should have %d arguments, got %d", narg, len(c.Input))
		}
		l1, err := deserialize("[]string", c.Input[0])
		if err != nil {
			return fmt.Errorf("cannot parse %s as []string", c.Input[0])
		}
		l2, err := goutils.SplitArray(c.Input[1])
		if err != nil {
			return fmt.Errorf("%s is not a valid list", c.Input[0])
		}
		if l1.Len() != len(l2) {
			return fmt.Errorf("input[0] and input[1] should have the same length")
		}
		if c.HasOutput() {
			l3, err := goutils.SplitArray(c.Output)
			if err != nil {
				return fmt.Errorf("%s is not a valid list", c.Input[0])
			}
			if l1.Len() != len(l3) {
				return fmt.Errorf("input and output should have the same length")
			}
		}
		return nil
	}

	resultType := q.MetaData.ResultType()
	if len(c.Input) != narg {
		return fmt.Errorf("should have %d arguments, got %d", narg, len(c.Input))
	}
	for j, arg := range c.Input {
		tp := q.MetaData.Params[j].Type
		if _, err := deserialize(tp, arg); err != nil {
			return fmt.Errorf("cannot parse %s as %s", arg, tp)
		}
	}
	if c.HasOutput() {
		if _, err := deserialize(resultType, c.Output); err != nil {
			return fmt.Errorf("cannot parse %s as %s", c.Output, resultType)
		}
	}

	return nil
}

func (c *TestCase) InputString() string {
	return utils.EnsureTrailingNewline(strings.Join(c.Input, "\n"))
}

func (c *TestCase) HasOutput() bool {
	return c.Output != ""
}

type TestCases struct {
	Cases    []TestCase
	Question *leetcode.QuestionData
}

func (tc *TestCases) AddCase(c TestCase) {
	c.No = len(tc.Cases) + 1
	c.Question = tc.Question
	tc.Cases = append(tc.Cases, c)
}

func (tc *TestCases) Contains(c TestCase) bool {
	for _, tc := range tc.Cases {
		if reflect.DeepEqual(c.Input, tc.Input) {
			return true
		}
	}
	return false
}

func (tc *TestCases) String() string {
	buf := new(bytes.Buffer)
	for i, c := range tc.Cases {
		buf.WriteString(testCaseInputMark + "\n")
		buf.WriteString(c.InputString())
		buf.WriteString(testCaseOutputMark + "\n")
		buf.WriteString(c.Output + "\n")
		if i != len(tc.Cases)-1 {
			buf.WriteString("\n")
		}
	}
	return buf.String()
}

func (tc *TestCases) InputString() string {
	buf := new(bytes.Buffer)
	for _, c := range tc.Cases {
		buf.WriteString(c.InputString())
	}
	return buf.String()
}

func (tc *TestCases) Check() error {
	for i, c := range tc.Cases {
		if err := c.Check(); err != nil {
			return fmt.Errorf("case %d: %w", i, err)
		}
	}
	return nil
}

// UpdateOutputs updates the output of the test cases with the result of the last run.
func (tc *TestCases) UpdateOutputs(answers []string) (bool, error) {
	if len(answers) != len(tc.Cases) {
		return false, fmt.Errorf("expected %d answers, got %d", len(tc.Cases), len(answers))
	}
	updated := false
	for i, c := range tc.Cases {
		if c.Output != answers[i] {
			c.Output = answers[i]
			tc.Cases[i] = c
			updated = true
		}
	}
	return updated, nil
}

func ParseTestCases(q *leetcode.QuestionData, f *FileOutput) (TestCases, error) {
	tc := TestCases{Question: q}

	content, err := f.GetContent()
	if err != nil {
		return tc, err
	}
	var (
		inputLines    []string
		output        string
		inputStarted  bool
		outputStarted bool
	)
	lines := utils.SplitLines(content)
	for _, line := range lines {
		line := strings.TrimSpace(line)
		switch {
		case line == "":
			continue
		case strings.HasPrefix(line, testCaseInputMark):
			inputStarted = true
			outputStarted = false
			if len(inputLines) > 0 {
				tc.AddCase(
					TestCase{
						Input:  slices.Clone(inputLines),
						Output: output,
					},
				)
				inputLines = inputLines[:0]
				output = ""
			}
		case strings.HasPrefix(line, testCaseOutputMark):
			outputStarted = true
			inputStarted = false
		case inputStarted:
			inputLines = append(inputLines, line)
		case outputStarted:
			if len(output) > 0 {
				return tc, errors.New("invalid test case: output should be a single line")
			}
			output = line
		}
	}
	if len(inputLines) > 0 {
		tc.AddCase(
			TestCase{
				Input:  slices.Clone(inputLines),
				Output: output,
			},
		)
	}

	if err := tc.Check(); err != nil {
		return tc, fmt.Errorf("invalid test case: %w", err)
	}

	return tc, nil
}

type Range struct {
	whole  bool
	max    int
	ranges [][2]int
}

func (r *Range) Contains(idx int) bool {
	if r.whole {
		return true
	}
	for _, rg := range r.ranges {
		if idx >= rg[0] && idx <= rg[1] {
			return true
		}
	}
	return false
}

func ParseRange(expr string, max int) (*Range, error) {
	r := &Range{max: max}

	if expr == "" || expr == "-" {
		r.whole = true
		return r, nil
	}

	parts := strings.Split(expr, ",")
	for _, part := range parts {
		var start, end int
		var startNegative bool
		if strings.HasPrefix(part, "-") {
			startNegative = true
			part = part[1:]
		}
		rangeParts := strings.SplitN(part, "-", 2)
		switch len(rangeParts) {
		case 1:
			idx, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid range: %s", part)
			}
			if startNegative && idx < 0 {
				return nil, fmt.Errorf("invalid range: %s", part)
			}
			if startNegative {
				idx = -idx
			}
			start, end = idx, idx
		case 2:
			var err error
			start, err = strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid range: %s", part)
			}
			if startNegative && start < 0 {
				return nil, fmt.Errorf("invalid range: %s", part)
			}
			if startNegative {
				start = -start
			}
			endStr := rangeParts[1]
			if endStr == "" {
				end = -1
			} else {
				end, err = strconv.Atoi(rangeParts[1])
				if err != nil {
					return nil, fmt.Errorf("invalid range: %s", part)
				}
			}
		default:
			return nil, fmt.Errorf("invalid range: %s", part)
		}

		if start < 0 {
			start = max + start + 1
		}
		if end < 0 {
			end = max + end + 1
		}
		if start <= 0 || start > max {
			return nil, fmt.Errorf("invalid range: %s", part)
		}
		if end <= 0 || end > max {
			return nil, fmt.Errorf("invalid range: %s", part)
		}
		if start > end {
			return nil, fmt.Errorf("invalid range: %s", part)
		}
		r.ranges = append(r.ranges, [2]int{start, end})
	}

	return r, nil
}
