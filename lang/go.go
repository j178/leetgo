package lang

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
)

const (
	goTestTemplate = `
func main() {
	// deserialize param
	// call function
    %s(%s, testcases)
    // get output param
    // serialize output and write to stdout
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
	modPath := "leetcode-solutions"
	var stderr bytes.Buffer
	cmd := exec.Command("go", "mod", "init", modPath)
	cmd.Dir = outDir
	cmd.Stdout = os.Stdout
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
	genResult, err := g.GeneratePaths(q)
	if err != nil {
		return false, fmt.Errorf("generate paths failed: %w", err)
	}
	genResult.SetOutDir(outDir)

	args := []string{"go", "run", "./" + genResult.SubDir}
	err = runTest(q, genResult, args, outDir)
	return err == nil, nil
}

func (g golang) generateTestContent(q *leetcode.QuestionData) (string, error) {
	var funcName, testFuncName string
	if q.MetaData.SystemDesign {
		funcName = "Constructor"
		testFuncName = "RunClassTestsWithString"
	} else {
		funcName = q.MetaData.Name
		testFuncName = "RunTestsWithString"
	}
	// TODO
	// TODO 根据 output.paramindex 找到真正的 output 对象
	testContent := fmt.Sprintf(
		goTestTemplate,
		testFuncName,
		funcName,
	)
	return testContent, nil
}

func (g golang) generateCodeFile(
	q *leetcode.QuestionData,
	filename string,
	blocks []config.Block,
	modifiers []ModifierFunc,
	separateDescriptionFile bool,
) (
	FileOutput,
	error,
) {
	testContent, err := g.generateTestContent(q)
	if err != nil {
		return FileOutput{}, err
	}
	blocks = append(
		blocks,
		config.Block{
			Name: "_internalBeforeMarker",
			Template: fmt.Sprintf(
				`package main

import . "%s"`, config.GoTestUtilsModPath,
			),
		},
		config.Block{
			Name:     "_internalAfterMarker",
			Template: testContent,
		},
	)
	content, err := g.generateCodeContent(
		q,
		blocks,
		modifiers,
		separateDescriptionFile,
	)
	if err != nil {
		return FileOutput{}, err
	}
	return FileOutput{
		Filename: filename,
		Content:  content,
		Type:     CodeFile | TestFile,
	}, nil
}

func (g golang) GeneratePaths(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, g)
	baseFilename, err := q.GetFormattedFilename(g.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	genResult := &GenerateResult{
		SubDir:   baseFilename,
		Question: q,
		Lang:     g,
	}
	genResult.AddFile(
		FileOutput{
			Filename: "solution.go",
			Type:     CodeFile | TestFile,
		},
	)
	genResult.AddFile(
		FileOutput{
			Filename: "testcases.txt",
			Type:     TestCasesFile,
		},
	)
	if separateDescriptionFile(g) {
		genResult.AddFile(
			FileOutput{
				Filename: "question.md",
				Type:     DocFile,
			},
		)
	}
	return genResult, nil
}

var goBuiltinModifiers = map[string]ModifierFunc{
	"removeUselessComments": removeUselessComments,
	"changeReceiverName":    changeReceiverName,
	"addNamedReturn":        addNamedReturn,
	"addMod":                addMod,
}

func (g golang) Generate(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, g)
	baseFilename, err := q.GetFormattedFilename(g.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	genResult := &GenerateResult{
		Question: q,
		Lang:     g,
		SubDir:   baseFilename,
	}

	separateDescriptionFile := separateDescriptionFile(g)
	blocks := getBlocks(g)
	modifiers, err := getModifiers(g, goBuiltinModifiers)
	if err != nil {
		return nil, err
	}
	codeFile, err := g.generateCodeFile(q, "solution.go", blocks, modifiers, separateDescriptionFile)
	if err != nil {
		return nil, err
	}
	testcaseFile, err := g.generateTestCasesFile(q, "testcases.txt")
	if err != nil {
		return nil, err
	}
	genResult.AddFile(codeFile)
	genResult.AddFile(testcaseFile)

	if separateDescriptionFile {
		docFile, err := g.generateDescriptionFile(q, "question.md")
		if err != nil {
			return nil, err
		}
		genResult.AddFile(docFile)
	}

	return genResult, nil
}
