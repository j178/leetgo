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
	requirements = "pytest>=7\n"
	confTest     = `import json


def _parse_test_cases(test_cases_str):
	inputs, outputs = [], []
	current = inputs
	for line in test_cases_str.splitlines():
		line = line.strip()
		if not line:
			if inputs:
				yield inputs, outputs
				inputs, outputs = [], []
		elif line.startswith("input:"):
			current = inputs
		elif line.startswith("output:"):
			current = outputs
		else:
			current.append(json.loads(line))
	if inputs:
		yield inputs, outputs


def pytest_generate_tests(metafunc):
	if "input_args" not in metafunc.fixturenames:
		return
	arg_names = ["input_args", "expected"]
	if "methods" in metafunc.fixturenames:
		arg_names.append("methods")
	matrix = []
	for inputs, outputs in _parse_test_cases(metafunc.module.TEST_CASES):
		if "methods" in metafunc.fixturenames:
			methods, inputs = inputs
			matrix.append([inputs, outputs[0], methods])
		else:
			# Assume there is only one output value
			matrix.append([inputs, outputs[0]])
	metafunc.parametrize(arg_names, matrix)
`
	pytestTemplate = `from solution import Solution

TEST_CASES = """\
%s
"""


def test_solution(input_args, expected):
	result = Solution().%s(*input_args)
	assert result == expected
`
	pytestSystemTemplate = `from solution import Solution

TEST_CASES = """\
%s
"""


def test_solution(input_args, expected, methods):
    assert methods[0] == "Solution"
    solution = Solution(*input_args[0])
    results = [None]
    for method, args in zip(methods[1:], input_args[1:]):
        result = getattr(solution, method)(*args)
        results.append(result)
    assert results == expected
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
	_, err = tryWrite(path.Join(outDir, "conftest.py"), confTest)
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
	if q.MetaData.SystemDesign {
		return fmt.Sprintf(pytestSystemTemplate, testcases)
	} else {
		return fmt.Sprintf(pytestTemplate, testcases, q.MetaData.Name)
	}
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
	var caseAndOutputs []string
	for i := 0; i < len(cases) && i/argsNum < len(outputs); i += argsNum {
		input := strings.Join(cases[i:i+argsNum], "\n")
		caseAndOutputs = append(
			caseAndOutputs,
			fmt.Sprintf("%s\n%s\n%s\n%s", testCaseInputMark, input, testCaseOutputMark, outputs[i/argsNum]),
		)
	}
	return strings.Join(caseAndOutputs, "\n\n")
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
	filenameTmpl := getFilenameTemplate(q, p)
	baseFilename, err := q.GetFormattedFilename(p.slug, filenameTmpl)
	if err != nil {
		return false, err
	}
	cmd := exec.Command(path.Join(outDir, ".venv", config.VenvPython), "-m", "pytest", baseFilename)
	cmd.Dir = outDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	return err == nil, nil
}
