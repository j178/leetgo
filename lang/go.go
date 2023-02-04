package lang

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
)

const (
	goTestFileTemplate = `
package main

import (
    "testing"

    . "%s"
)

var testcases = ` + "`" + `
%s
` + "`" + `

func Test_%s(t *testing.T) {
    targetCaseNum := 0
    // targetCaseNum := -1
    if err := %s(t, %s, testcases, targetCaseNum); err != nil {
        t.Fatal(err)
    }
}
`
)

type golang struct {
	baseLang
}

func addNamedReturn(code string, q *leetcode.QuestionData) string {
	lines := strings.Split(code, "\n")
	var newLines []string
	skipNext := 0
	for _, line := range lines {
		if skipNext > 0 {
			skipNext--
			continue
		}
		if strings.HasPrefix(line, "func ") {
			rightBrace := strings.LastIndex(line, ")")
			returnType := strings.TrimSpace(line[rightBrace+1 : strings.LastIndex(line, "{")])
			if returnType != "" {
				if returnType == "bool" || returnType == "string" {
					newLines = append(newLines, line)
				} else if q.MetaData.SystemDesign && strings.Contains(line, "func Constructor") {
					newLines = append(newLines, line)
					newLines = append(newLines, "\n\treturn "+returnType+"{}")
					skipNext = 1
				} else {
					newLines = append(newLines, line[:rightBrace+1]+" (ans "+returnType+") {")
					newLines = append(newLines, "\n\treturn")
					skipNext = 1
				}
			} else {
				newLines = append(newLines, line)
			}
		} else {
			newLines = append(newLines, line)
		}
	}
	return strings.Join(newLines, "\n")
}

func changeReceiverName(code string, q *leetcode.QuestionData) string {
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "func (this *") {
			n := len("func (this *")
			prefix := strings.ToLower(line[n : n+1])
			lines[i] = strings.Replace(line, "this", prefix, 1)
		}
	}
	return strings.Join(lines, "\n")
}

func addMod(code string, q *leetcode.QuestionData) string {
	if q.MetaData.SystemDesign {
		return code
	}
	content, _ := q.GetContent()
	if !needsMod(content) {
		return code
	}

	lines := strings.Split(code, "\n")
	var newLines []string
	var returnType string
	for _, line := range lines {
		if strings.HasPrefix(line, "func ") {
			if strings.Count(line, "(") == 1 {
				rightBrace := strings.LastIndex(line, ")")
				returnType = strings.TrimSpace(line[rightBrace+1 : strings.LastIndex(line, "{")])
			} else {
				s := line[strings.LastIndex(line, "(")+1 : strings.LastIndex(line, ")")]
				returnType = s[strings.LastIndex(s, " ")+1:]
			}
			newLines = append(newLines, line)
			newLines = append(newLines, "\tconst mod int = 1e9 + 7\n")
		} else if strings.HasPrefix(line, "\treturn") {
			if returnType == "int" || returnType == "int64" || returnType == "int32" {
				newLines = append(newLines, "\tans = (ans%mod + mod) % mod")
			}
			newLines = append(newLines, line)
		} else {
			newLines = append(newLines, line)
		}
	}
	return strings.Join(newLines, "\n")
}

func (g golang) HasInitialized(outDir string) (bool, error) {
	cmd := exec.Command("go", "list", "-m", "-json", config.GoTestUtilsModPath)
	cmd.Dir = outDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		if bytes.Contains(output, []byte("not a known dependency")) || bytes.Contains(
			output,
			[]byte("go.mod file not found"),
		) {
			return false, nil
		}
		return false, fmt.Errorf("go list failed: %w", err)
	}
	return true, nil
}

func (g golang) Initialize(outDir string) error {
	modPath := config.Get().Code.Go.GoModPath
	if modPath == "" {
		modPath = "leetcode-solutions"
		hclog.L().Warn("`code.go.go_mod_path` is not set, use default path", "mod_path", modPath)
	}
	var stderr bytes.Buffer
	cmd := exec.Command("go", "mod", "init", modPath)
	cmd.Dir = outDir
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	err := cmd.Run()
	if err != nil && !bytes.Contains(stderr.Bytes(), []byte("go.mod already exists")) {
		return err
	}

	cmd = exec.Command("go", "get", config.GoTestUtilsModPath)
	cmd.Dir = outDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
}

func (g golang) RunLocalTest(q *leetcode.QuestionData, outDir string) (bool, error) {
	cmd := exec.Command("go", "list", "-m")
	cmd.Dir = outDir
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("go list failed: %w", err)
	}
	modPath := strings.TrimSpace(string(output))
	if modPath == "" {
		return false, fmt.Errorf("go mod path is empty")
	}

	genResult, err := g.GeneratePaths(q)
	if err != nil {
		return false, fmt.Errorf("generate paths failed: %w", err)
	}
	path := genResult.Files[0].Path
	basePath := filepath.Clean(filepath.Dir(path))

	cmd = exec.Command("go", "test", "-v", modPath+"/"+basePath)
	cmd.Dir = outDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	return err == nil, nil
}

func (g golang) generateTest(q *leetcode.QuestionData, testcases string) string {
	var funcName, testFuncName string
	if q.MetaData.SystemDesign {
		funcName = "Constructor"
		testFuncName = "RunClassTestsWithString"
	} else {
		funcName = q.MetaData.Name
		testFuncName = "RunTestsWithString"
	}
	content := fmt.Sprintf(testFileHeader, g.lineComment)
	content += fmt.Sprintf(goTestFileTemplate, config.GoTestUtilsModPath, testcases, funcName, testFuncName, funcName)
	return content
}

func (g golang) GeneratePaths(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, g)
	baseFilename, err := q.GetFormattedFilename(g.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	codeFile := filepath.Join(baseFilename, "solution.go")
	testFile := filepath.Join(baseFilename, "solution_test.go")

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
		Lang:     g,
		Files:    files,
	}, nil
}

var goBuiltinModifiers = map[string]ModifierFunc{
	"removeUselessComments": removeUselessComments,
	"changeReceiverName":    changeReceiverName,
	"addNamedReturn":        addNamedReturn,
	"addMod":                addMod,
}

func (g golang) Generate(q *leetcode.QuestionData) (*GenerateResult, error) {
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
	codeFile := filepath.Join(baseFilename, "solution.go")
	testFile := filepath.Join(baseFilename, "solution_test.go")

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
