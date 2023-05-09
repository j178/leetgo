package lang

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/constants"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

var requirements = fmt.Sprintf("sortedcontainers\n%s\n", constants.PythonTestUtilsMode)

type python struct {
	baseLang
}

func (p python) Initialize(outDir string) error {
	pythonExe := config.Get().Code.Python.Executable

	cmd := exec.Command(pythonExe, "--version")
	log.Info("checking python version", "cmd", cmd.String())
	versionOutput, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	pythonVersion := strings.TrimPrefix(string(versionOutput), "Python ")
	if !strings.HasPrefix(pythonVersion, "3.") {
		return errors.New("python version must be 3.x")
	}

	err = utils.WriteFile(path.Join(outDir, "requirements.txt"), []byte(requirements))
	if err != nil {
		return err
	}

	cmd = exec.Command(pythonExe, "-m", "venv", ".venv")
	log.Info("creating venv", "cmd", cmd.String())
	cmd.Dir = outDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return err
	}

	_ = utils.WriteFile(path.Join(outDir, ".venv", ".gitignore"), []byte("*\n"))

	cmd = exec.Command(
		path.Join(outDir, ".venv", constants.VenvPython),
		"-m",
		"pip",
		"install",
		"--disable-pip-version-check",
		"-Ur",
		"requirements.txt",
	)
	log.Info("pip install", "cmd", cmd.String())
	cmd.Dir = outDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
}

func (p python) HasInitialized(outDir string) (bool, error) {
	return utils.IsExist(path.Join(outDir, ".venv")), nil
}

func (p python) RunLocalTest(q *leetcode.QuestionData, outDir string, targetCase string) (bool, error) {
	genResult, err := p.GeneratePaths(q)
	if err != nil {
		return false, err
	}
	genResult.SetOutDir(outDir)

	testFile := genResult.GetFile(TestFile).GetPath()
	cmd := []string{path.Join(outDir, ".venv", constants.VenvPython), testFile}
	return runTest(q, genResult, cmd, targetCase)
}

func toPythonType(typeName string) string {
	switch typeName {
	case "integer":
		return "int"
	case "string":
		return "str"
	case "long":
		return "int"
	case "double":
		return "float"
	case "boolean":
		return "bool"
	case "character":
		return "str"
	case "void":
		return ""
	case "TreeNode":
		return "TreeNode"
	case "ListNode":
		return "ListNode"
	default:
		if strings.HasSuffix(typeName, "[]") {
			return "List[" + toPythonType(typeName[:len(typeName)-2]) + "]"
		}
	}
	return typeName
}

func (p python) generateNormalTestCode(q *leetcode.QuestionData) (string, error) {
	const template = `if __name__ == "__main__":
%s
`
	code := ""
	paramNames := make([]string, 0, len(q.MetaData.Params))
	for _, param := range q.MetaData.Params {
		varName := param.Name
		varType := toPythonType(param.Type)
		code += fmt.Sprintf(
			"\t%s: %s = deserialize(\"%s\", read_line())\n",
			varName,
			varType,
			varType,
		)
		if !param.HelperParam {
			paramNames = append(paramNames, param.Name)
		}
	}
	if q.MetaData.Return != nil && q.MetaData.Return.Type != "void" {
		code += fmt.Sprintf(
			"\tans = Solution().%s(%s)\n",
			q.MetaData.Name,
			strings.Join(paramNames, ", "),
		)
	} else {
		code += fmt.Sprintf(
			"\t%s(%s)\n",
			q.MetaData.Name,
			strings.Join(paramNames, ", "),
		)
		ansName := paramNames[q.MetaData.Output.ParamIndex]
		code += fmt.Sprintf("\tans = %s\n", ansName)
	}
	code += fmt.Sprintf(
		"\tprint(\"%s\", serialize(ans))",
		testCaseOutputMark,
	)

	testContent := fmt.Sprintf(template, code)
	return testContent, nil
}

func (p python) generateSystemDesignTestCode(q *leetcode.QuestionData) (string, error) {
	const template = `if __name__ == "__main__":
	ops: List[str] = deserialize("List[str]", read_line())
	params = split_array(read_line())
	output = ["null"]

%s

	for i in range(1, len(ops)):
		match ops[i]:
%s

	print("%s " + join_array(output))
`
	var prepareCode string
	paramNames := make([]string, 0, len(q.MetaData.Constructor.Params))
	if len(q.MetaData.Constructor.Params) > 0 {
		prepareCode += "\tconstructor_params = split_array(params[0])\n"
		for i, param := range q.MetaData.Constructor.Params {
			varName := param.Name
			varType := toPythonType(param.Type)
			prepareCode += fmt.Sprintf(
				"\t%s: %s = deserialize(\"%s\", constructor_params[%d])\n",
				varName,
				varType,
				varType,
				i,
			)
			paramNames = append(paramNames, varName)
		}
	}
	prepareCode += fmt.Sprintf(
		"\tobj = %s(%s)",
		q.MetaData.ClassName,
		strings.Join(paramNames, ", "),
	)

	callCode := ""
	for _, method := range q.MetaData.Methods {
		methodCall := "\t\t\tcase \"" + method.Name + "\":\n"
		if len(method.Params) > 0 {
			methodCall += "\t\t\t\tmethod_params = split_array(params[i])\n"
		}
		methodParamNames := make([]string, 0, len(method.Params))
		for i, param := range method.Params {
			varName := param.Name
			varType := toPythonType(param.Type)
			methodCall += fmt.Sprintf(
				"\t\t\t\t%s: %s = deserialize(\"%s\", method_params[%d])\n",
				varName,
				varType,
				varType,
				i,
			)
			methodParamNames = append(methodParamNames, varName)
		}
		if method.Return.Type != "" && method.Return.Type != "void" {
			methodCall += fmt.Sprintf(
				"\t\t\t\tans = serialize(obj.%s(%s))\n\t\t\t\toutput.append(ans)\n",
				method.Name,
				strings.Join(methodParamNames, ", "),
			)
		} else {
			methodCall += fmt.Sprintf(
				"\t\t\t\tobj.%s(%s)\n",
				toPythonType(method.Name),
				strings.Join(methodParamNames, ", "),
			)
			methodCall += "\t\t\t\toutput.append(\"null\")\n"
		}
		callCode += methodCall
	}
	callCode = callCode[:len(callCode)-1] // remove last newline
	testContent := fmt.Sprintf(
		template,
		prepareCode,
		callCode,
		testCaseOutputMark,
	)
	return testContent, nil
}

func (p python) generateTestContent(q *leetcode.QuestionData) (string, error) {
	if q.MetaData.SystemDesign {
		return p.generateSystemDesignTestCode(q)
	}
	return p.generateNormalTestCode(q)
}

func (p python) generateCodeFile(
	q *leetcode.QuestionData,
	filename string,
	blocks []config.Block,
	modifiers []ModifierFunc,
	separateDescriptionFile bool,
) (
	FileOutput,
	error,
) {
	codeHeader := fmt.Sprintf(
		`from typing import *
from %s import *`, constants.PythonTestUtilsMode,
	)
	testContent, err := p.generateTestContent(q)
	if err != nil {
		return FileOutput{}, err
	}
	blocks = append(
		[]config.Block{
			{
				Name:     beforeBeforeMarker,
				Template: codeHeader,
			},
			{
				Name:     afterAfterMarker,
				Template: testContent,
			},
		},
		blocks...,
	)
	content, err := p.generateCodeContent(
		q,
		blocks,
		modifiers,
		separateDescriptionFile,
	)
	if err != nil {
		return FileOutput{}, err
	}
	content = strings.ReplaceAll(content, "\t", "    ")
	return FileOutput{
		Filename: filename,
		Content:  content,
		Type:     CodeFile | TestFile,
	}, nil
}

func (p python) GeneratePaths(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, p)
	baseFilename, err := q.GetFormattedFilename(p.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	genResult := &GenerateResult{
		SubDir:   baseFilename,
		Question: q,
		Lang:     p,
	}
	genResult.AddFile(
		FileOutput{
			Filename: "solution.py",
			Type:     CodeFile | TestFile,
		},
	)
	genResult.AddFile(
		FileOutput{
			Filename: "testcases.txt",
			Type:     TestCasesFile,
		},
	)
	if separateDescriptionFile(p) {
		genResult.AddFile(
			FileOutput{
				Filename: "question.md",
				Type:     DocFile,
			},
		)
	}
	return genResult, nil
}

func (p python) Generate(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, p)
	baseFilename, err := q.GetFormattedFilename(p.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	genResult := &GenerateResult{
		Question: q,
		Lang:     p,
		SubDir:   baseFilename,
	}

	separateDescriptionFile := separateDescriptionFile(p)
	blocks := getBlocks(p)
	modifiers, err := getModifiers(p, builtinModifiers)
	if err != nil {
		return nil, err
	}
	codeFile, err := p.generateCodeFile(q, "solution.py", blocks, modifiers, separateDescriptionFile)
	if err != nil {
		return nil, err
	}
	testcaseFile, err := p.generateTestCasesFile(q, "testcases.txt")
	if err != nil {
		return nil, err
	}
	genResult.AddFile(codeFile)
	genResult.AddFile(testcaseFile)

	if separateDescriptionFile {
		docFile, err := p.generateDescriptionFile(q, "question.md")
		if err != nil {
			return nil, err
		}
		genResult.AddFile(docFile)
	}

	return genResult, nil
}
