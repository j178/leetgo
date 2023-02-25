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

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	goutils "github.com/j178/leetgo/testutils/go"
	"github.com/j178/leetgo/utils"
	"github.com/jedib0t/go-pretty/v6/list"
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
	goTpName := convertToGoType(tpName)
	ty := typeNameToType(goTpName)
	if ty == nil {
		return reflect.Value{}, fmt.Errorf("unknown type: %s", tpName)
	}
	return goutils.DeserializeValue(ty, raw)
}

type testCase struct {
	no          int
	input       []string
	output      string
	outputValue reflect.Value
}

func (c testCase) Input() string {
	return utils.EnsureTrailingNewline(strings.Join(c.input, "\n"))
}

type testCases struct {
	cases      []testCase
	targetCase int
}

func checkTestCases(q *leetcode.QuestionData, tc testCases) error {
	narg := q.MetaData.NArg()
	resultType := q.MetaData.ResultType()
	for i, c := range tc.cases {
		if len(c.input) != narg {
			return fmt.Errorf("should have %d arguments, got %d", narg, len(c.input))
		}
		for j, arg := range c.input {
			tp := q.MetaData.Params[j].Type
			if _, err := deserialize(tp, arg); err != nil {
				return fmt.Errorf("cannot parse %s as %s", arg, tp)
			}
		}
		if v, err := deserialize(resultType, c.output); err != nil {
			return fmt.Errorf("cannot parse %s as %s", c.output, resultType)
		} else {
			tc.cases[i].outputValue = v
		}
	}
	return nil
}

func parseTestCases(q *leetcode.QuestionData, f *FileOutput) (testCases, error) {
	tc := testCases{}
	content, err := f.GetContent()
	if err != nil {
		return tc, err
	}
	var (
		input         []string
		output        string
		inputStarted  bool
		outputStarted bool
	)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		switch {
		case strings.TrimSpace(line) == "":
			continue
		case strings.HasPrefix(line, testCaseTargetMark):
			no := strings.TrimSpace(line[len(testCaseTargetMark):])
			targetCase, err := strconv.Atoi(no)
			if err != nil {
				return tc, fmt.Errorf("invalid target_case: %s is not valid number", no)
			}
			tc.targetCase = targetCase
		case strings.HasPrefix(line, testCaseInputMark):
			inputStarted = true
			outputStarted = false
			if len(input) > 0 && len(output) > 0 {
				tc.cases = append(
					tc.cases, testCase{
						no:     len(tc.cases) + 1,
						input:  append([]string(nil), input...),
						output: output,
					},
				)
				input = input[:0]
				output = ""
			}
		case strings.HasPrefix(line, testCaseOutputMark):
			outputStarted = true
			inputStarted = false
		case inputStarted:
			input = append(input, line)
		case outputStarted:
			if len(output) > 0 {
				return tc, errors.New("invalid test case: output should be a single line")
			}
			output = line
		}
	}
	if len(input) > 0 && len(output) > 0 {
		tc.cases = append(
			tc.cases, testCase{
				no:     len(tc.cases) + 1,
				input:  append([]string(nil), input...),
				output: output,
			},
		)
	}
	if tc.targetCase > len(tc.cases) {
		return tc, fmt.Errorf("invalid target_case: %d, maximum is %d", tc.targetCase, len(tc.cases))
	}
	if tc.targetCase < 0 {
		tc.targetCase += len(tc.cases)
	}

	if err := checkTestCases(q, tc); err != nil {
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

func parseOutput(q *leetcode.QuestionData, outputLine string) (reflect.Value, error) {
	if outputLine == "" {
		return reflect.Value{}, fmt.Errorf("no output found")
	}
	tp := q.MetaData.ResultType()
	v, err := deserialize(tp, outputLine)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("invalid output: %w", err)
	}
	return v, nil
}

func judgeResult(q *leetcode.QuestionData, actual, expected reflect.Value) bool {
	// TODO compare by question rules
	return reflect.DeepEqual(actual.Interface(), expected.Interface())
}

func runTest(q *leetcode.QuestionData, genResult *GenerateResult, args []string, outDir string) (bool, error) {
	testcaseFile := genResult.GetFile(TestCasesFile)
	if testcaseFile == nil {
		panic("no test cases file generated")
	}
	tc, err := parseTestCases(q, testcaseFile)
	if err != nil {
		return false, err
	}
	if len(tc.cases) == 0 {
		return false, fmt.Errorf("no test cases found")
	}
	var (
		outputBuf bytes.Buffer
		ran       int
		passed    int
	)
	l := list.NewWriter()
	l.SetStyle(list.StyleBulletCircle)

	for _, c := range tc.cases {
		func() {
			if tc.targetCase != 0 && c.no != tc.targetCase {
				l.AppendItem(fmt.Sprintf("Case %d:    Skipped", c.no))
				return
			}
			ran++
			outputBuf.Reset()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, args[0], args[1:]...)
			cmd.Dir = outDir
			cmd.Stdin = strings.NewReader(c.Input())
			cmd.Stdout = &outputBuf
			cmd.Stderr = &outputBuf
			err = cmd.Start()
			if err != nil {
				l.AppendItem(fmt.Sprintf("Case %d:    Failed to start", c.no))
				return
			}
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()
			select {
			case <-ctx.Done():
			case err = <-done:
			}

			actualOutput, stdout := extractOutput(outputBuf.String())
			if ctx.Err() != nil {
				l.AppendItem(fmt.Sprintf("Case %d:    Time limit exceeded", c.no))
				l.Indent()
				l.AppendItem(fmt.Sprintf("Input:      %s", strings.ReplaceAll(c.Input(), "\n", "↩ ")))
				if stdout != "" {
					l.AppendItem(fmt.Sprintf("Stdout:     %s", strings.ReplaceAll(stdout, "\n", "↩ ")))
				}
				l.UnIndent()
				return
			}
			if err != nil {
				l.AppendItem(fmt.Sprintf("Case %d:    Runtime error", c.no))
				l.Indent()
				l.AppendItem(fmt.Sprintf("Input:      %s", strings.ReplaceAll(c.Input(), "\n", "↩ ")))
				if stdout != "" {
					l.AppendItem(fmt.Sprintf("Stdout:     %s", strings.ReplaceAll(stdout, "\n", "↩ ")))
				}
				l.UnIndent()
				return
			}
			actualOutputValue, err := parseOutput(q, actualOutput)
			if err != nil {
				l.AppendItem(fmt.Sprintf("Case %d:    Invalid output", c.no))
				l.Indent()
				l.AppendItem(fmt.Sprintf("Input:      %s", strings.ReplaceAll(c.Input(), "\n", "↩ ")))
				l.AppendItem(fmt.Sprintf("Output:     %s", actualOutput))
				if stdout != "" {
					l.AppendItem(fmt.Sprintf("Stdout:     %s", strings.ReplaceAll(stdout, "\n", "↩ ")))
				}
				l.UnIndent()
				return
			}

			if judgeResult(q, actualOutputValue, c.outputValue) {
				passed++
				l.AppendItem(fmt.Sprintf("Case %d:    Accepted", c.no))
			} else {
				l.AppendItem(fmt.Sprintf("Case %d:    Wrong answer", c.no))
				l.Indent()
				l.AppendItem(fmt.Sprintf("Input:      %s", strings.ReplaceAll(c.Input(), "\n", "↩ ")))
				l.AppendItem(fmt.Sprintf("Output:     %s", actualOutput))
				l.AppendItem(fmt.Sprintf("Expected:   %s", c.output))
				if stdout != "" {
					l.AppendItem(fmt.Sprintf("Stdout:     %s", strings.ReplaceAll(stdout, "\n", "↩ ")))
				}
				l.UnIndent()
			}
		}()
	}
	fmt.Println(l.Render())

	if passed == ran {
		return true, nil
	}
	return false, nil
}
