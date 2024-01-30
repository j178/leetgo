package lang

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

const leetgoGo = "github.com/j178/leetgo/testutils/go"

var goDeps = []string{
	leetgoGo + "@v0.2.0",
}

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
	content, _ := q.GetPreferContent()
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

func (g golang) shouldInit(outDir string) (bool, error) {
	if !utils.IsExist(filepath.Join(outDir, "go.mod")) {
		return true, nil
	}

	update, err := IsDepUpdateToDate(g)
	if err != nil {
		return false, err
	}
	if !update {
		return true, nil
	}
	return false, nil
}

func (g golang) InitWorkspace(outDir string) error {
	if should, err := g.shouldInit(outDir); err != nil || !should {
		return err
	}

	err := utils.RemoveIfExist(filepath.Join(outDir, "go.mod"))
	if err != nil {
		return err
	}
	_ = utils.RemoveIfExist(filepath.Join(outDir, "go.sum"))

	const modPath = "leetcode-solutions"
	var stderr strings.Builder
	cmd := exec.Command("go", "mod", "init", modPath)
	log.Info("go mod init", "cmd", cmd.String())
	cmd.Dir = outDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	err = cmd.Run()
	if err != nil && !strings.Contains(stderr.String(), "go.mod already exists") {
		return err
	}

	cmd = exec.Command("go", "get")
	cmd.Args = append(cmd.Args, goDeps...)
	log.Info("go get", "cmd", cmd.String())
	cmd.Dir = outDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	err = UpdateDep(g)
	return err
}

func (g golang) RunLocalTest(q *leetcode.QuestionData, outDir string, targetCase string) (bool, error) {
	genResult, err := g.GeneratePaths(q)
	if err != nil {
		return false, fmt.Errorf("generate paths failed: %w", err)
	}
	genResult.SetOutDir(outDir)

	testFile := genResult.GetFile(TestFile).GetPath()
	if !utils.IsExist(testFile) {
		return false, fmt.Errorf("file %s not found", utils.RelToCwd(testFile))
	}
	execFile, err := getTempBinFile(q, g)
	if err != nil {
		return false, fmt.Errorf("get temp bin file failed: %w", err)
	}

	err = buildTest(q, genResult, []string{"go", "build", "-o", execFile, testFile})
	if err != nil {
		return false, fmt.Errorf("build failed: %w", err)
	}

	return runTest(q, genResult, []string{execFile}, targetCase)
}

// toGoType converts LeetCode type name to Go type name.
func toGoType(typeName string) string {
	switch typeName {
	case "integer":
		return "int"
	case "long":
		return "int64"
	case "double":
		return "float64"
	case "boolean":
		return "bool"
	case "character":
		return "byte"
	case "void":
		return ""
	case "TreeNode":
		return "*TreeNode"
	case "ListNode":
		return "*ListNode"
	default:
		if strings.HasSuffix(typeName, "[]") {
			return "[]" + toGoType(typeName[:len(typeName)-2])
		}
	}
	return typeName
}

// nolint: staticcheck
func toGoFuncName(f string) string {
	return strings.Title(f)
}

func (g golang) generateNormalTestCode(q *leetcode.QuestionData) (string, error) {
	const template = `func main() {
	stdin := bufio.NewReader(os.Stdin)
%s
	fmt.Println("\n%s", Serialize(ans))
}
`
	code := ""
	paramNames := make([]string, 0, len(q.MetaData.Params))
	for _, param := range q.MetaData.Params {
		code += fmt.Sprintf(
			"\t%s := Deserialize[%s](ReadLine(stdin))\n",
			param.Name,
			toGoType(param.Type),
		)
		paramNames = append(paramNames, param.Name)
	}
	if q.MetaData.Return != nil && q.MetaData.Return.Type != "void" {
		code += fmt.Sprintf(
			"\tans := %s(%s)\n",
			q.MetaData.Name,
			strings.Join(paramNames, ", "),
		)
	} else {
		code += fmt.Sprintf(
			"\t%s(%s)\n",
			q.MetaData.Name,
			strings.Join(paramNames, ", "),
		)
		if q.MetaData.Output != nil {
			ansName := paramNames[q.MetaData.Output.ParamIndex]
			code += fmt.Sprintf("\tans := %s\n", ansName)
		} else {
			code += "\tans := \"\"\n"
		}
	}

	testContent := fmt.Sprintf(template, code, testCaseOutputMark)

	if q.MetaData.Manual {
		testContent = fmt.Sprintf("// %s\n%s", manualWarning, testContent)
	}
	return testContent, nil
}

func (g golang) generateSystemDesignTestCode(q *leetcode.QuestionData) (string, error) {
	const template = `func main() {
	stdin := bufio.NewReader(os.Stdin)
	ops := Deserialize[[]string](ReadLine(stdin))
	params := MustSplitArray(ReadLine(stdin))
	output := make([]string, 0, len(ops))
	output = append(output, "null")

%s

	for i := 1; i < len(ops); i++ {
		switch ops[i] {
%s
		}
	}
	fmt.Println("\n%s", JoinArray(output))
}
`
	var prepareCode string
	var paramNames []string
	if len(q.MetaData.Constructor.Params) > 0 {
		prepareCode += "\tconstructorParams := MustSplitArray(params[0])\n"
		for i, param := range q.MetaData.Constructor.Params {
			prepareCode += fmt.Sprintf(
				"\t%s := Deserialize[%s](constructorParams[%d])\n",
				param.Name,
				toGoType(param.Type),
				i,
			)
			paramNames = append(paramNames, param.Name)
		}
	}
	prepareCode += fmt.Sprintf("\tobj := Constructor(%s)", strings.Join(paramNames, ", "))

	callCode := ""
	for _, method := range q.MetaData.Methods {
		methodCall := "\t\tcase \"" + method.Name + "\":\n"
		if len(method.Params) > 0 {
			methodCall += "\t\t\tmethodParams := MustSplitArray(params[i])\n"
		}
		var methodParamNames []string
		for i, param := range method.Params {
			methodCall += fmt.Sprintf(
				"\t\t\t%s := Deserialize[%s](methodParams[%d])\n",
				param.Name,
				toGoType(param.Type),
				i,
			)
			methodParamNames = append(methodParamNames, param.Name)
		}
		if method.Return.Type != "" && method.Return.Type != "void" {
			methodCall += fmt.Sprintf(
				"\t\t\tans := Serialize(obj.%s(%s))\n\t\t\toutput = append(output, ans)\n",
				toGoFuncName(method.Name),
				strings.Join(methodParamNames, ", "),
			)
		} else {
			methodCall += fmt.Sprintf(
				"\t\t\tobj.%s(%s)\n",
				toGoFuncName(method.Name),
				strings.Join(methodParamNames, ", "),
			)
			methodCall += "\t\t\toutput = append(output, \"null\")\n"
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

	if q.MetaData.Manual {
		testContent = fmt.Sprintf("// %s\n%s", manualWarning, testContent)
	}
	return testContent, nil
}

func (g golang) generateTestContent(q *leetcode.QuestionData) (string, error) {
	if q.MetaData.SystemDesign {
		return g.generateSystemDesignTestCode(q)
	}
	return g.generateNormalTestCode(q)
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
	codeHeader := fmt.Sprintf(
		`package main

import (
	"bufio"
	"fmt"
	"os"

	. "%s"
)`, leetgoGo,
	)
	testContent, err := g.generateTestContent(q)
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
