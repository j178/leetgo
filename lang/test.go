package lang

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"

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
			return fmt.Errorf("invalid number of arguments in case %d", c.no)
		}
		for j, arg := range c.input {
			if _, err := deserialize(q.MetaData.Params[j].Type, arg); err != nil {
				return fmt.Errorf("invalid argument in case %d: %w", c.no, err)
			}
		}
		if v, err := deserialize(resultType, c.output); err != nil {
			return fmt.Errorf("invalid result in case %d: %w", c.no, err)
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
			targetCase, err := strconv.Atoi(strings.TrimSpace(line[len(testCaseTargetMark):]))
			if err != nil {
				return tc, fmt.Errorf("invalid target_case: %w", err)
			}
			tc.targetCase = targetCase
		case strings.HasPrefix(line, testCaseInputMark):
			inputStarted = true
			outputStarted = false
			if len(input) > 0 && len(output) > 0 {
				tc.cases = append(
					tc.cases, testCase{
						no:     len(tc.cases),
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
			output = line
		}
	}
	if len(input) > 0 && len(output) > 0 {
		tc.cases = append(
			tc.cases, testCase{
				no:     len(tc.cases),
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
		return tc, err
	}

	return tc, nil
}

func parseOutput(q *leetcode.QuestionData, out string) (reflect.Value, error) {
	var outputLine string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, testCaseOutputMark) {
			outputLine = strings.TrimSpace(line[len(testCaseOutputMark):])
			break
		}
	}
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

func judgeResult(q *leetcode.QuestionData, output, expected reflect.Value) bool {
	// TODO compare by question rules
	return reflect.DeepEqual(output.Interface(), expected.Interface())
}

func runTest(q *leetcode.QuestionData, genResult *GenerateResult, args []string, outDir string) error {
	testcaseFile := genResult.GetFile(TestCasesFile)
	if testcaseFile == nil {
		panic("no test cases file generated")
	}
	tc, err := parseTestCases(q, testcaseFile)
	if err != nil {
		return err
	}
	var (
		outputBuf bytes.Buffer
		passed    int
	)
	for _, c := range tc.cases {
		if tc.targetCase != 0 && c.no != tc.targetCase {
			continue
		}
		outputBuf.Reset()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Dir = outDir
		cmd.Stdin = strings.NewReader(c.Input())
		cmd.Stdout = io.MultiWriter(&outputBuf, os.Stdout)
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			// todo show error
			cancel()
			continue
		}

		output, err := parseOutput(q, outputBuf.String())
		if err != nil {
			// show error
			cancel()
			continue
		}

		if judgeResult(q, output, c.outputValue) {
			passed++
		}
		cancel()
	}

	return nil
}
