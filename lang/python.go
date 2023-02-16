package lang

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
)

const (
	requirements   = "pytest>=7\n"
	pytestTemplate = `import pytest

from solution import Solution

TEST_CASES = %s


@pytest.mark.parametrize("input_args,expected", TEST_CASES)
def test_solution(input_args, expected):
	result = Solution().%s(*input_args)
	assert result == expected
`
)

type python struct {
	baseLang
}

func (p python) Initialize(outDir string) error {
	python := config.Get().Code.Python.PythonExecutable

	pythonExe, err := exec.LookPath(python)

	if err != nil {
		return fmt.Errorf("python executable %v not found in PATH", python)
	}

	_, err = tryWrite(path.Join(outDir, "requirements.txt"), requirements)
	if err != nil {
		return err
	}
	cmd := exec.Command(pythonExe, "-m", "venv", path.Join(outDir, ".venv"))
	if err = cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command(path.Join(outDir, ".venv", config.VenvPython), "-m", "pip", "install", "-Ur", "requirements.txt")
	cmd.Dir = outDir
	err = cmd.Run()
	return err
}

func (p python) HasInitialized(outDir string) (bool, error) {
	_, err := os.Stat(path.Join(outDir, ".venv"))
	return !os.IsNotExist(err), nil
}

func (g python) Generate(q *leetcode.QuestionData) (*GenerateResult, error) {
	blocks := getBlocks(g)
	modifiers, err := getModifiers(g, goBuiltinModifiers)
	if err != nil {
		return nil, err
	}
	codeContent, err := g.generateContent(q, blocks, modifiers)
	if err != nil {
		return nil, err
	}

	testcaseStr := g.generateTestCases(q)
	testContent := g.generateTest(q, testcaseStr)

	filenameTmpl := getFilenameTemplate(q, g)
	baseFilename, err := q.GetFormattedFilename(g.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	codeFile := filepath.Join(baseFilename, "solution.py")
	testFile := filepath.Join(baseFilename, "solution_test.py")

	files := []FileOutput{
		{
			Path:    codeFile,
			Content: codeContent,
			Type:    CodeFile,
		},
		{
			Path:    testFile,
			Content: testContent,
			Type:    TestFile,
		},
	}

	return &GenerateResult{
		Question: q,
		Lang:     g,
		Files:    files,
	}, nil
}

func (p python) generateTest(q *leetcode.QuestionData, testcases string) string {
	return fmt.Sprintf(pytestTemplate, testcases, q.MetaData.Name)
}

func (p python) generateTestCases(q *leetcode.QuestionData) string {
	cases := q.GetTestCases()
	outputs := q.ParseExampleOutputs()
	argsNum := 0
	if q.MetaData.SystemDesign {
		argsNum = 2
	} else {
		argsNum = len(q.MetaData.Params)
	}

	// Assume all questions output are single.
	caseAndOutputs := []string{"["}
	for i := 0; i < len(cases) && i/argsNum < len(outputs); i += argsNum {
		input := strings.Join(cases[i:i+argsNum], ", ")
		caseAndOutputs = append(
			caseAndOutputs,
			fmt.Sprintf("    ([%s], %s),", input, outputs[i/argsNum]),
		)
	}
	caseAndOutputs = append(caseAndOutputs, "]")
	return strings.Join(caseAndOutputs, "\n")
}

func (p python) GeneratePaths(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, p)
	baseFilename, err := q.GetFormattedFilename(p.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	codeFile := filepath.Join(baseFilename, "solution.py")
	testFile := filepath.Join(baseFilename, "solution_test.py")

	files := []FileOutput{
		{
			Path: codeFile,
			Type: CodeFile,
		},
		{
			Path: testFile,
			Type: TestFile,
		},
	}

	return &GenerateResult{
		Question: q,
		Lang:     p,
		Files:    files,
	}, nil
}

func (p python) RunLocalTest(q *leetcode.QuestionData, outDir string) (bool, error) {
	cmd := exec.Command(path.Join(outDir, ".venv", config.VenvPython), "-m", "pytest")
	cmd.Dir = outDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	return err == nil, nil
}
