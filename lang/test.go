package lang

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/jedib0t/go-pretty/v6/list"

	goutils "github.com/j178/leetgo/testutils/go"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

func RunLocalTest(q *leetcode.QuestionData, targetCase string) (bool, error) {
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

	return tester.RunLocalTest(q, outDir, targetCase)
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

// deserialize deserializes a string to a reflect.Value. The tpName is a LeeCode type name.
func deserialize(tpName string, raw string) (reflect.Value, error) {
	raw = strings.TrimSpace(raw)
	goTpName := toGoType(tpName)
	ty := typeNameToType(goTpName)
	if ty == nil {
		return reflect.Value{}, fmt.Errorf("unknown type: %s", tpName)
	}
	return goutils.DeserializeValue(ty, raw)
}

// extractOutput extracts the output from the stdout of the test program.
func extractOutput(s string) (string, string) {
	var output string
	var others []string
	for _, line := range utils.SplitLines(s) {
		if strings.HasPrefix(line, testCaseOutputMark) {
			// If there are multiple output lines, only the last one is used.
			output = strings.TrimSpace(line[len(testCaseOutputMark):])
		} else {
			others = append(others, line)
		}
	}
	return output, strings.Join(others, "\n")
}

// checkOutput checks if the output is valid.
// If the question is a system design question, it checks if the output is an array and has the same length as the input.
// Otherwise, it checks if the output can be deserialized to the expected type.
func checkOutput(q *leetcode.QuestionData, input []string, outputLine string) error {
	if outputLine == "" {
		return fmt.Errorf("no output found")
	}
	if q.MetaData.SystemDesign {
		arr, err := goutils.SplitArray(outputLine)
		if err != nil {
			return fmt.Errorf("invalid output: %s", outputLine)
		}
		inputArr, _ := goutils.SplitArray(input[0])
		if len(arr) != len(inputArr) {
			return fmt.Errorf("output length mismatch: expected %d, got %d", len(inputArr), len(arr))
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

func buildTest(_ *leetcode.QuestionData, genResult *GenerateResult, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	buf := new(bytes.Buffer)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = genResult.OutDir
	cmd.Stdout = buf
	cmd.Stderr = buf

	testFile := genResult.GetFile(TestFile).GetPath()
	if log.GetLevel() <= log.DebugLevel {
		log.Info("building", "cmd", cmd.String())
	} else {
		log.Info("building", "file", utils.RelToCwd(testFile))
	}
	err := cmd.Run()
	if err != nil {
		fmt.Println(config.StdoutStyle.Render(strings.TrimSuffix(buf.String(), "\n")))
		return err
	}
	return nil
}

func runTest(q *leetcode.QuestionData, genResult *GenerateResult, args []string, targetCaseStr string) (bool, error) {
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
	caseRange, err := ParseRange(targetCaseStr, len(tc.Cases))
	if err != nil {
		return false, err
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
			if !caseRange.Contains(c.No) {
				l.AppendItem(fmt.Sprintf("Case %d:    %s", c.No, config.SkippedStyle.Render("Skipped")))
				return
			}
			if !c.HasOutput() {
				l.AppendItem(fmt.Sprintf("Case %d:    %s", c.No, config.SkippedStyle.Render("Skipped: no output")))
				return
			}
			ran++
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			outputBuf := new(bytes.Buffer)
			cmd := exec.CommandContext(ctx, args[0], args[1:]...)
			cmd.Dir = genResult.OutDir
			cmd.Stdin = strings.NewReader(c.InputString())
			cmd.Stdout = outputBuf
			cmd.Stderr = outputBuf
			err = cmd.Start()
			if err != nil {
				l.AppendItem(fmt.Sprintf("Case %d:    %s", c.No, config.ErrorStyle.Render("Failed to start")))
				return
			}
			err = cmd.Wait()

			actualOutput, stdout := extractOutput(outputBuf.String())
			mayAppendStdout := func() {
				if stdout != "" {
					out := config.StdoutStyle.Render(utils.TruncateString(stdout, 1000))
					l.AppendItem(fmt.Sprintf("Stdout:     %s", out))
				}
			}
			if ctx.Err() != nil {
				l.AppendItem(fmt.Sprintf("Case %d:    %s", c.No, config.ErrorStyle.Render("Time limit exceeded")))
				l.Indent()
				l.AppendItem(
					fmt.Sprintf(
						"Input:      %s",
						utils.TruncateString(strings.ReplaceAll(c.InputString(), "\n", "↩ "), 100),
					),
				)
				mayAppendStdout()
				l.UnIndent()
				return
			}
			if err != nil {
				l.AppendItem(fmt.Sprintf("Case %d:    %s", c.No, config.ErrorStyle.Render("Runtime error")))
				l.Indent()
				l.AppendItem(
					fmt.Sprintf(
						"Input:      %s",
						utils.TruncateString(strings.ReplaceAll(c.InputString(), "\n", "↩ "), 100),
					),
				)
				mayAppendStdout()
				l.UnIndent()
				return
			}
			err = checkOutput(q, c.Input, actualOutput)
			if err != nil {
				l.AppendItem(fmt.Sprintf("Case %d:    %s", c.No, config.ErrorStyle.Render("Invalid output")))
				l.Indent()
				l.AppendItem(
					fmt.Sprintf(
						"Input:      %s",
						utils.TruncateString(strings.ReplaceAll(c.InputString(), "\n", "↩ "), 100),
					),
				)
				l.AppendItem(fmt.Sprintf("Output:     %s", utils.TruncateString(actualOutput, 100)))
				mayAppendStdout()
				l.UnIndent()
				return
			}

			if r := judger.Judge(c.Input, c.Output, actualOutput); r.IsAccepted() {
				passed++
				l.AppendItem(fmt.Sprintf("Case %d:    %s", c.No, config.PassedStyle.Render("Passed")))
			} else {
				l.AppendItem(fmt.Sprintf("Case %d:    %s", c.No, config.FailedStyle.Render("Wrong answer")))
				l.Indent()
				l.AppendItem(fmt.Sprintf("Reason:     %s", r.GetInfo()))
				l.AppendItem(
					fmt.Sprintf(
						"Input:      %s",
						utils.TruncateString(strings.ReplaceAll(c.InputString(), "\n", "↩ "), 100),
					),
				)
				l.AppendItem(fmt.Sprintf("Output:     %s", utils.TruncateString(actualOutput, 100)))
				l.AppendItem(fmt.Sprintf("Expected:   %s", utils.TruncateString(c.Output, 100)))
				mayAppendStdout()
				l.UnIndent()
			}
		}()
	}

	return passed == ran, nil
}
