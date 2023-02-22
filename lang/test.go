package lang

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
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

type testCase struct {
	no     int
	input  string
	output string
}

type testCases struct {
	cases      []testCase
	targetCase int
}

func parseTestCases(f *FileOutput) (testCases, error) {
	tc := testCases{}
	content, err := f.GetContent()
	if err != nil {
		return tc, err
	}
	lines := strings.Split(content, "\n")
	var input, output []string
	var inputStarted, outputStarted bool
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
		case strings.HasPrefix(line, "input:"):
			inputStarted = true
			outputStarted = false
			if len(input) > 0 && len(output) > 0 {
				tc.cases = append(
					tc.cases, testCase{
						no:     len(tc.cases),
						input:  strings.Join(input, "\n"),
						output: strings.Join(output, "\n"),
					},
				)
				input = input[:0]
				output = output[:0]
			}
		case strings.HasPrefix(line, "output:"):
			outputStarted = true
			inputStarted = false
		case inputStarted:
			input = append(input, line)
		case outputStarted:
			output = append(output, line)
		}
	}
	if len(input) > 0 && len(output) > 0 {
		tc.cases = append(
			tc.cases, testCase{
				no:     len(tc.cases),
				input:  strings.Join(input, "\n"),
				output: strings.Join(output, "\n"),
			},
		)
	}
	if tc.targetCase > len(tc.cases) {
		return tc, fmt.Errorf("invalid target_case: %d, maximum is %d", tc.targetCase, len(tc.cases))
	}
	if tc.targetCase < 0 {
		tc.targetCase += len(tc.cases)
	}
	// TODO check parameters count and deserialize
	return tc, nil
}

func parseOutput(out string) string {
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, testCaseTargetMark) {
			return strings.TrimSpace(line[len(testCaseTargetMark):])
		}
	}
	return ""
}

func judgeOutput(output, expected string) bool {
	return output == expected
}

func runTest(q *leetcode.QuestionData, genResult *GenerateResult, args []string, outDir string) error {
	testcaseFile := genResult.GetFile(TestCasesFile)
	if testcaseFile == nil {
		panic("no test cases file generated")
	}
	tc, err := parseTestCases(testcaseFile)
	if err != nil {
		return err
	}
	var (
		outputBuf bytes.Buffer
		passed    int
	)
	for _, testcase := range tc.cases {
		if tc.targetCase != 0 && testcase.no != tc.targetCase {
			continue
		}
		outputBuf.Reset()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Dir = outDir
		cmd.Stdin = strings.NewReader(testcase.input)
		cmd.Stdout = io.MultiWriter(&outputBuf, os.Stdout)
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			// todo show error
			continue
		}

		output := parseOutput(outputBuf.String())
		if output == "" {
			// show error
			continue
		}
		if judgeOutput(output, testcase.output) {
			passed++
		}
	}

	return nil
}
