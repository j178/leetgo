package lang

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jedib0t/go-pretty/v6/list"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	goutils "github.com/j178/leetgo/testutils/go"
	"github.com/j178/leetgo/utils"
)

func RunLocalTest(q *leetcode.QuestionData) (bool, error) {
	cfg := config.Get()
	gen, err := GetGenerator(cfg.Code.Lang)
	if err != nil {
		return false, err
	}
	tester, ok := gen.(LocalTestable)
	if !ok {
		return false, fmt.Errorf("language %s does not support local test", gen.Slug())
	}
	err = q.Fulfill()
	if err != nil {
		return false, fmt.Errorf("failed to get question data: %w", err)
	}
	outDir := getOutDir(q, gen)
	if !utils.IsExist(outDir) {
		return false, fmt.Errorf("no code generated for %s in language %s", q.TitleSlug, gen.Slug())
	}

	return tester.RunLocalTest(q, outDir)
}

// typeNameToType converts a Go type name to reflect.Type.
func typeNameToType(ty string) reflect.Type {
	switch ty {
	case "int":
		return reflect.TypeOf(0)
	case "int64":
		return reflect.TypeOf(int64(0))
	case "float64":
		return reflect.TypeOf(float64(0))
	case "string":
		return reflect.TypeOf("")
	case "bool":
		return reflect.TypeOf(false)
	case "byte":
		return reflect.TypeOf(byte(0))
	case "*TreeNode":
		return reflect.TypeOf((*goutils.TreeNode)(nil))
	case "*ListNode":
		return reflect.TypeOf((*goutils.ListNode)(nil))
	default:
		if strings.HasPrefix(ty, "[]") {
			et := typeNameToType(ty[2:])
			if et == nil {
				return nil
			}
			return reflect.SliceOf(et)
		}
	}
	return nil
}

func deserialize(tpName string, raw string) (reflect.Value, error) {
	raw = strings.TrimSpace(raw)
	goTpName := toGoType(tpName)
	ty := typeNameToType(goTpName)
	if ty == nil {
		return reflect.Value{}, fmt.Errorf("unknown type: %s", tpName)
	}
	return goutils.DeserializeValue(ty, raw)
}

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
		l3, err := goutils.SplitArray(c.Output)
		if err != nil {
			return fmt.Errorf("%s is not a valid list", c.Input[0])
		}
		if l1.Len() != len(l2) || l1.Len() != len(l3) {
			return fmt.Errorf("Input and output should have the same length")
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
	if _, err := deserialize(resultType, c.Output); err != nil {
		return fmt.Errorf("cannot parse %s as %s", c.Output, resultType)
	}
	return nil
}

func (c *TestCase) InputString() string {
	return utils.EnsureTrailingNewline(strings.Join(c.Input, "\n"))
}

type TestCases struct {
	TargetCase int
	Cases      []TestCase
	Question   *leetcode.QuestionData
}

func (tc *TestCases) AddCase(c TestCase) {
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
	_, _ = fmt.Fprintf(buf, "%s %d\n\n", testCaseTargetMark, tc.TargetCase)
	for i, c := range tc.Cases {
		_, _ = fmt.Fprintln(buf, testCaseInputMark)
		_, _ = fmt.Fprint(buf, c.InputString())
		_, _ = fmt.Fprintln(buf, testCaseOutputMark)
		_, _ = fmt.Fprintln(buf, c.Output)
		if i != len(tc.Cases)-1 {
			_, _ = fmt.Fprintln(buf)
		}
	}
	return buf.String()
}

func (tc *TestCases) Check() error {
	q := tc.Question
	err := q.Fulfill()
	if err != nil {
		return fmt.Errorf("failed to get question data: %w", err)
	}

	if tc.TargetCase < 0 || tc.TargetCase >= len(tc.Cases) {
		return fmt.Errorf("target_case %d is out of range", tc.TargetCase)
	}
	for i, c := range tc.Cases {
		if err := c.Check(); err != nil {
			return fmt.Errorf("case %d: %w", i, err)
		}
	}
	return nil
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
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line := strings.TrimSpace(line)
		switch {
		case line == "":
			continue
		case strings.HasPrefix(line, testCaseTargetMark):
			no := strings.TrimSpace(line[len(testCaseTargetMark):])
			targetCase, err := strconv.Atoi(no)
			if err != nil {
				return tc, fmt.Errorf("invalid target_case: %s is not valid number", no)
			}
			tc.TargetCase = targetCase
		case strings.HasPrefix(line, testCaseInputMark):
			inputStarted = true
			outputStarted = false
			if len(inputLines) > 0 && len(output) > 0 {
				tc.Cases = append(
					tc.Cases, TestCase{
						Question: q,
						No:       len(tc.Cases) + 1,
						Input:    append([]string(nil), inputLines...),
						Output:   output,
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
	if len(inputLines) > 0 && len(output) > 0 {
		tc.Cases = append(
			tc.Cases, TestCase{
				Question: q,
				No:       len(tc.Cases) + 1,
				Input:    append([]string(nil), inputLines...),
				Output:   output,
			},
		)
	}
	if tc.TargetCase < 0 {
		tc.TargetCase += len(tc.Cases) + 1
	}

	if err := tc.Check(); err != nil {
		return tc, fmt.Errorf("invalid test case: %w", err)
	}

	return tc, nil
}

func extractOutput(s string) (string, string) {
	var output string
	var others []string
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(line, testCaseOutputMark) {
			// If there are multiple output lines, only the last one is used.
			output = strings.TrimSpace(line[len(testCaseOutputMark):])
		} else {
			others = append(others, line)
		}
	}
	return output, strings.Join(others, "\n")
}

func checkOutput(q *leetcode.QuestionData, outputLine string) error {
	if outputLine == "" {
		return fmt.Errorf("no output found")
	}
	if q.MetaData.SystemDesign {
		_, err := goutils.SplitArray(outputLine)
		if err != nil {
			return fmt.Errorf("invalid output: %s", outputLine)
		}
		return nil
	}
	tp := q.MetaData.ResultType()
	_, err := deserialize(tp, outputLine)
	if err != nil {
		return fmt.Errorf("invalid output: %s", outputLine)
	}
	return nil
}

var (
	skippedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#b8b8b8"))
	passedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00b300"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	failedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6600"))
	stdoutStyle  = lipgloss.NewStyle().Faint(true)
)

func runTest(q *leetcode.QuestionData, genResult *GenerateResult, args []string, outDir string) (bool, error) {
	testcaseFile := genResult.GetFile(TestCasesFile)
	if testcaseFile == nil {
		panic("no test cases file generated")
	}
	tc, err := ParseTestCases(q, testcaseFile)
	if err != nil {
		return false, err
	}
	if len(tc.Cases) == 0 {
		return false, fmt.Errorf("no test cases found")
	}

	judger := GetJudger(q)

	var ran, passed int
	for _, c := range tc.Cases {
		func() {
			l := list.NewWriter()
			l.SetStyle(list.StyleBulletCircle)
			defer func() {
				fmt.Println(l.Render())
			}()
			if tc.TargetCase != 0 && c.No != tc.TargetCase {
				l.AppendItem(fmt.Sprintf("Case %d:    %s", c.No, skippedStyle.Render("Skipped")))
				return
			}
			ran++
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			outputBuf := new(bytes.Buffer)
			cmd := exec.CommandContext(ctx, args[0], args[1:]...)
			cmd.Dir = outDir
			cmd.Stdin = strings.NewReader(c.InputString())
			cmd.Stdout = outputBuf
			cmd.Stderr = outputBuf
			err = cmd.Start()
			if err != nil {
				l.AppendItem(fmt.Sprintf("Case %d:    %s", c.No, errorStyle.Render("Failed to start")))
				return
			}
			err = cmd.Wait()

			actualOutput, stdout := extractOutput(outputBuf.String())
			mayAppendStdout := func() {
				if stdout != "" {
					out := stdoutStyle.Render(strings.ReplaceAll(stdout, "\n", "↩ "))
					l.AppendItem(fmt.Sprintf("Stdout:     %s", out))
				}
			}
			if ctx.Err() != nil {
				l.AppendItem(fmt.Sprintf("Case %d:    %s", c.No, errorStyle.Render("Time limit exceeded")))
				l.Indent()
				l.AppendItem(fmt.Sprintf("Input:      %s", strings.ReplaceAll(c.InputString(), "\n", "↩ ")))
				mayAppendStdout()
				l.UnIndent()
				return
			}
			if err != nil {
				l.AppendItem(fmt.Sprintf("Case %d:    %s", c.No, errorStyle.Render("Runtime error")))
				l.Indent()
				l.AppendItem(fmt.Sprintf("Input:      %s", strings.ReplaceAll(c.InputString(), "\n", "↩ ")))
				mayAppendStdout()
				l.UnIndent()
				return
			}
			err = checkOutput(q, actualOutput)
			if err != nil {
				l.AppendItem(fmt.Sprintf("Case %d:    %s", c.No, errorStyle.Render("Invalid output")))
				l.Indent()
				l.AppendItem(fmt.Sprintf("Input:      %s", strings.ReplaceAll(c.InputString(), "\n", "↩ ")))
				l.AppendItem(fmt.Sprintf("Output:     %s", actualOutput))
				mayAppendStdout()
				l.UnIndent()
				return
			}

			if judger.Judge(actualOutput, c.Output) {
				passed++
				l.AppendItem(fmt.Sprintf("Case %d:    %s", c.No, passedStyle.Render("Accepted")))
			} else {
				l.AppendItem(fmt.Sprintf("Case %d:    %s", c.No, failedStyle.Render("Wrong answer")))
				l.Indent()
				l.AppendItem(fmt.Sprintf("Input:      %s", strings.ReplaceAll(c.InputString(), "\n", "↩ ")))
				l.AppendItem(fmt.Sprintf("Output:     %s", actualOutput))
				l.AppendItem(fmt.Sprintf("Expected:   %s", c.Output))
				mayAppendStdout()
				l.UnIndent()
			}
		}()
	}
	if passed == ran {
		return true, nil
	}
	return false, nil
}
