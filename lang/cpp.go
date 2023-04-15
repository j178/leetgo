package lang

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	cppUtils "github.com/j178/leetgo/testutils/cpp"
	"github.com/j178/leetgo/utils"
)

type cpp struct {
	baseLang
}

func (c cpp) Initialize(outDir string) error {
	headerPath := filepath.Join(outDir, cppUtils.HeaderName)
	if _, err := tryWrite(headerPath, cppUtils.HeaderContent); err != nil {
		return err
	}
	return nil
}

func (c cpp) HasInitialized(outDir string) (bool, error) {
	headerPath := filepath.Join(outDir, cppUtils.HeaderName)
	return utils.IsExist(headerPath), nil
}

var cppTypes = map[string]string{
	"void":      "void",
	"integer":   "int",
	"long":      "int64_t",
	"string":    "string",
	"double":    "double",
	"ListNode":  "ListNode*",
	"TreeNode":  "TreeNode*",
	"boolean":   "bool",
	"character": "char",
}

const (
	objectName                 = "obj"
	returnName                 = "res"
	inputStreamName            = "cin"
	outputStreamName           = "out_stream"
	systemDesignMethodMapName  = "methods"
	systemDesignMethodNameName = "method_name"
	systemDesignMethodListName = "method_names"
)

func (c cpp) getCppTypeName(t string) (int, string) {
	return strings.Count(t, "[]"), cppTypes[strings.ReplaceAll(t, "[]", "")]
}

func (c cpp) getVectorTypeName(d int, t string) string {
	return strings.Repeat("vector<", d) + t + strings.Repeat(">", d)
}

func (c cpp) getDeclCodeForType(d int, t string, n string) string {
	return fmt.Sprintf("%s %s;", c.getVectorTypeName(d, t), n)
}

func (c cpp) getScanCodeForType(n string, ifs string) string {
	return fmt.Sprintf("LeetCodeIO::scan(%s, %s);", ifs, n)
}

func (c cpp) getPrintCodeForType(n string, ofs string) string {
	/* assumes one invocation for each printed variable */
	/* (parameter "n" could be a function call, which we only wish to call once) */
	return fmt.Sprintf("LeetCodeIO::print(%s, %s);", ofs, n)
}

func (c cpp) getParamString(params []leetcode.MetaDataParam) string {
	var paramList []string
	for _, param := range params {
		paramList = append(paramList, param.Name)
	}
	return strings.Join(paramList, ", ")
}

func (c cpp) generateScanCode(q *leetcode.QuestionData) string {
	if q.MetaData.SystemDesign {
		return fmt.Sprintf(
			"\t%s %s\n",
			c.getDeclCodeForType(1, "string", systemDesignMethodListName),
			c.getScanCodeForType(systemDesignMethodListName, inputStreamName),
		)
	}

	var scanCode string
	for _, param := range q.MetaData.Params {
		dimCnt, cppType := c.getCppTypeName(param.Type)
		scanCode += fmt.Sprintf(
			"\t%s %s\n",
			c.getDeclCodeForType(dimCnt, cppType, param.Name),
			c.getScanCodeForType(param.Name, inputStreamName),
		)
	}
	return scanCode
}

func (c cpp) generateInitCode(q *leetcode.QuestionData) string {
	if q.MetaData.SystemDesign {
		return fmt.Sprintf("\t%s *%s;\n", q.MetaData.ClassName, objectName)
	} else {
		return fmt.Sprintf("\tSolution *%s = new Solution();\n", objectName)
	}
}

func (c cpp) generateCallCode(q *leetcode.QuestionData) (callCode string) {
	generateParamScanningCode := func(params []leetcode.MetaDataParam) {
		if len(params) > 0 {
			for _, param := range params {
				dimCnt, cppType := c.getCppTypeName(param.Type)
				callCode += fmt.Sprintf(
					"\t\t\t%s %s %s.ignore();\n",
					c.getDeclCodeForType(dimCnt, cppType, param.Name),
					c.getScanCodeForType(param.Name, inputStreamName),
					inputStreamName,
				)
			}
		} else {
			callCode += fmt.Sprintf("\t\t\t%s.ignore();\n", inputStreamName)
		}
	}

	if !q.MetaData.SystemDesign {
		callCode = fmt.Sprintf(
			"\tauto %s = %s->%s(%s);\n",
			returnName,
			objectName,
			q.MetaData.Name,
			c.getParamString(q.MetaData.Params),
		)
	} else {
		/* define methods */ {
			callCode = fmt.Sprintf(
				"\tconst unordered_map<string, function<void()>> %s = {\n",
				systemDesignMethodMapName,
			)
			/* operations in constructor function call */ {
				callCode += fmt.Sprintf("\t\t{ \"%s\", [&]() {\n", q.MetaData.ClassName)
				generateParamScanningCode(q.MetaData.Constructor.Params)
				callCode += fmt.Sprintf(
					"\t\t\t%s = new %s(%s);\n",
					objectName,
					q.MetaData.ClassName,
					c.getParamString(q.MetaData.Constructor.Params),
				)
				callCode += fmt.Sprintf("\t\t\t%s << \"null,\";\n\t\t} },\n", outputStreamName)
			}
			/* operations in member function calls */
			for _, method := range q.MetaData.Methods {
				callCode += fmt.Sprintf("\t\t{ \"%s\", [&]() {\n", method.Name)
				generateParamScanningCode(method.Params)
				_, returnType := c.getCppTypeName(method.Return.Type)
				functionCall := fmt.Sprintf(
					"%s->%s(%s)",
					objectName,
					method.Name,
					c.getParamString(method.Params),
				)
				if returnType != "void" {
					callCode += fmt.Sprintf(
						"\t\t\t%s %s << ',';\n",
						c.getPrintCodeForType(functionCall, outputStreamName),
						outputStreamName,
					)
				} else {
					callCode += fmt.Sprintf(
						"\t\t\t%s;\n\t\t\t%s << \"null,\";\n",
						functionCall,
						outputStreamName,
					)
				}
				callCode += "\t\t} },\n"
			}
			callCode += "\t};"
		}
		/* invoke methods */ {
			callCode += fmt.Sprintf(
				`
	%s << '[';
	for (auto &&%s : %s) {
		%s.ignore(2);
		%s.at(%s)();
	}
	%s.ignore();
	%s.seekp(-1, ios_base::end); %s << ']';
`,
				outputStreamName,
				systemDesignMethodNameName,
				systemDesignMethodListName,
				inputStreamName,
				systemDesignMethodMapName,
				systemDesignMethodNameName,
				inputStreamName,
				outputStreamName,
				outputStreamName,
			)
		}
	}
	return
}

func (c cpp) generatePrintCode(q *leetcode.QuestionData) (printCode string) {
	if !q.MetaData.SystemDesign {
		printCode += "\t" + c.getPrintCodeForType(returnName, outputStreamName) + "\n"
	}
	printCode += fmt.Sprintf("\tcout << \"%s \" << %s.rdbuf();\n", testCaseOutputMark, outputStreamName)
	return
}

func (c cpp) generateTestContent(q *leetcode.QuestionData) (string, error) {
	const template = `int main() {
	ios_base::sync_with_stdio(false);
	stringstream ` + outputStreamName + `;

%s

%s

%s

%s

	delete ` + objectName + `;
	return 0;
}`
	return fmt.Sprintf(
		template,
		c.generateScanCode(q),
		c.generateInitCode(q),
		c.generateCallCode(q),
		c.generatePrintCode(q),
	), nil
}

func usingGCC() bool {
	cxx := config.Get().Code.Cpp.CXX
	cmd := exec.Command(cxx, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return true
	}
	return !strings.Contains(string(output), "clang")
}

func (c cpp) generateCodeFile(
	q *leetcode.QuestionData,
	filename string,
	blocks []config.Block,
	modifiers []ModifierFunc,
	separateDescriptionFile bool,
) (
	FileOutput,
	error,
) {
	var stdCppHeader string
	if usingGCC() {
		stdCppHeader = "#include <bits/stdc++.h>\n"
	}

	codeHeader := fmt.Sprintf(
		`%s#include "%s"
using namespace std;

`,
		stdCppHeader,
		cppUtils.HeaderName,
	)

	testContent, err := c.generateTestContent(q)
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
	content, err := c.generateCodeContent(
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

func (c cpp) RunLocalTest(q *leetcode.QuestionData, outDir string) (bool, error) {
	genResult, err := c.GeneratePaths(q)
	if err != nil {
		return false, fmt.Errorf("generate paths failed: %w", err)
	}
	genResult.SetOutDir(outDir)

	testFile := genResult.GetFile(TestFile).GetPath()
	execFile, err := getTempBinFile(q, c)
	if err != nil {
		return false, fmt.Errorf("generate temporary binary file path failed: %w", err)
	}

	cfg := config.Get()
	compiler := cfg.Code.Cpp.CXX
	compilerFlags := append([]string(nil), cfg.Code.Cpp.CXXFLAGS...)
	compilerFlags = append(compilerFlags, "-I", outDir, "-o", execFile, testFile)

	cmd := exec.Command(compiler, compilerFlags...)
	cmd.Dir = outDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Info("compiling", "cmd", cmd.String())
	err = cmd.Run()
	if err != nil {
		return false, fmt.Errorf("compilation failed: %w", err)
	}

	return runTest(q, genResult, []string{execFile}, outDir)
}

func (c cpp) Generate(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, c)
	baseFilename, err := q.GetFormattedFilename(c.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	genResult := &GenerateResult{
		Question: q,
		Lang:     c,
		SubDir:   baseFilename,
	}

	separateDescriptionFile := separateDescriptionFile(c)
	blocks := getBlocks(c)
	modifiers, err := getModifiers(c, builtinModifiers)
	if err != nil {
		return nil, err
	}
	codeFile, err := c.generateCodeFile(q, "solution.cpp", blocks, modifiers, separateDescriptionFile)
	if err != nil {
		return nil, err
	}
	testcaseFile, err := c.generateTestCasesFile(q, "testcases.txt")
	if err != nil {
		return nil, err
	}
	genResult.AddFile(codeFile)
	genResult.AddFile(testcaseFile)

	if separateDescriptionFile {
		docFile, err := c.generateDescriptionFile(q, "question.md")
		if err != nil {
			return nil, err
		}
		genResult.AddFile(docFile)
	}

	return genResult, nil
}

func (c cpp) GeneratePaths(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, c)
	baseFilename, err := q.GetFormattedFilename(c.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	genResult := &GenerateResult{
		SubDir:   baseFilename,
		Question: q,
		Lang:     c,
	}
	genResult.AddFile(
		FileOutput{
			Filename: "solution.cpp",
			Type:     CodeFile | TestFile,
		},
	)
	genResult.AddFile(
		FileOutput{
			Filename: "testcases.txt",
			Type:     TestCasesFile,
		},
	)
	if separateDescriptionFile(c) {
		genResult.AddFile(
			FileOutput{
				Filename: "question.md",
				Type:     DocFile,
			},
		)
	}
	return genResult, nil
}
