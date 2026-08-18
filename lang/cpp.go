package lang

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/shlex"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	cppUtils "github.com/j178/leetgo/testutils/cpp"
	"github.com/j178/leetgo/utils"
)

type cpp struct {
	baseLang
}

func (c cpp) InitWorkspace(outDir string) error {
	// Generated code may use new LC_IO APIs, so always keep the managed header in sync.
	headerPath := filepath.Join(outDir, cppUtils.HeaderName)
	if err := utils.WriteFile(headerPath, cppUtils.HeaderContent); err != nil {
		return err
	}

	if should, err := c.shouldUpdateStdCxx(outDir); err != nil || !should {
		return err
	}

	stdCxxPath := filepath.Join(outDir, "bits", "stdc++.h")
	if err := utils.WriteFile(stdCxxPath, cppUtils.StdCxxContent); err != nil {
		return err
	}

	return UpdateDep(c)
}

func (c cpp) shouldUpdateStdCxx(outDir string) (bool, error) {
	stdCxxPath := filepath.Join(outDir, "bits", "stdc++.h")
	if !utils.IsExist(stdCxxPath) {
		return true, nil
	}

	upToDate, err := IsDepUpdateToDate(c)
	if err != nil {
		return false, err
	}
	return !upToDate, nil
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
	systemDesignMethodListName = "method_names"
	systemDesignOutputName     = "output"
)

func (c cpp) getCppTypeName(t string) (int, string) {
	return strings.Count(t, "[]"), cppTypes[strings.ReplaceAll(t, "[]", "")]
}

func (c cpp) getVectorTypeName(d int, t string) string {
	return strings.Repeat("vector<", d) + t + strings.Repeat(">", d)
}

func (c cpp) getDeserializeCodeForType(d int, t string, n string, input string) string {
	typeName := c.getVectorTypeName(d, t)
	return fmt.Sprintf("%s %s = LeetCodeIO::deserialize<%s>(%s);", typeName, n, typeName, input)
}

func (c cpp) getPrintCodeForType(n string, ofs string) string {
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
			"\tauto %s = LeetCodeIO::deserialize<vector<string>>(%s);\n"+
				"\tauto params = LeetCodeIO::split_array(%s);\n",
			systemDesignMethodListName,
			inputStreamName,
			inputStreamName,
		)
	}

	var scanCode string
	for _, param := range q.MetaData.Params {
		dimCnt, cppType := c.getCppTypeName(param.Type)
		scanCode += "\t" + c.getDeserializeCodeForType(
			dimCnt,
			cppType,
			param.Name,
			inputStreamName,
		) + "\n"
	}
	return scanCode
}

func (c cpp) generateInitCode(q *leetcode.QuestionData) string {
	if q.MetaData.SystemDesign {
		return fmt.Sprintf("\tunique_ptr<%s> %s;\n", q.MetaData.ClassName, objectName)
	}
	return fmt.Sprintf("\tSolution %s;\n", objectName)
}

func (c cpp) generateCallCode(q *leetcode.QuestionData) string {
	if !q.MetaData.SystemDesign {
		if q.MetaData.Return != nil && q.MetaData.Return.Type != "void" {
			return fmt.Sprintf(
				"\tauto %s = %s.%s(%s);\n",
				returnName,
				objectName,
				q.MetaData.Name,
				c.getParamString(q.MetaData.Params),
			)
		}
		return fmt.Sprintf(
			"\t%s.%s(%s);\n",
			objectName,
			q.MetaData.Name,
			c.getParamString(q.MetaData.Params),
		)
	}

	generateParamScanningCode := func(params []leetcode.MetaDataParam) string {
		var code string
		for index, param := range params {
			dimCnt, cppType := c.getCppTypeName(param.Type)
			code += "\t\t\t" + c.getDeserializeCodeForType(
				dimCnt,
				cppType,
				param.Name,
				fmt.Sprintf("method_params, %d", index),
			) + "\n"
		}
		return code
	}
	lambdaParameter := func(params []leetcode.MetaDataParam) string {
		if len(params) == 0 {
			return "const vector<string> &"
		}
		return "const vector<string> &method_params"
	}

	callCode := fmt.Sprintf(
		"\tvector<string> %s;\n\tconst unordered_map<string, function<void(const vector<string> &)>> %s = {\n",
		systemDesignOutputName,
		systemDesignMethodMapName,
	)

	callCode += fmt.Sprintf(
		"\t\t{ \"%s\", [&](%s) {\n",
		q.MetaData.ClassName,
		lambdaParameter(q.MetaData.Constructor.Params),
	)
	callCode += generateParamScanningCode(q.MetaData.Constructor.Params)
	callCode += fmt.Sprintf(
		"\t\t\t%s = make_unique<%s>(%s);\n",
		objectName,
		q.MetaData.ClassName,
		c.getParamString(q.MetaData.Constructor.Params),
	)
	callCode += fmt.Sprintf("\t\t\t%s.push_back(\"null\");\n\t\t} },\n", systemDesignOutputName)

	for _, method := range q.MetaData.Methods {
		callCode += fmt.Sprintf(
			"\t\t{ \"%s\", [&](%s) {\n",
			method.Name,
			lambdaParameter(method.Params),
		)
		callCode += generateParamScanningCode(method.Params)
		functionCall := fmt.Sprintf(
			"%s->%s(%s)",
			objectName,
			method.Name,
			c.getParamString(method.Params),
		)
		if method.Return.Type != "" && method.Return.Type != "void" {
			callCode += fmt.Sprintf(
				"\t\t\t%s.push_back(LeetCodeIO::serialize(%s));\n",
				systemDesignOutputName,
				functionCall,
			)
		} else {
			callCode += fmt.Sprintf(
				"\t\t\t%s;\n\t\t\t%s.push_back(\"null\");\n",
				functionCall,
				systemDesignOutputName,
			)
		}
		callCode += "\t\t} },\n"
	}
	callCode += "\t};\n"

	callCode += fmt.Sprintf(
		`	if (%s.size() != params.size()) {
		throw LeetCodeIO::Error("method and parameter counts differ");
	}
	for (size_t i = 0; i < %s.size(); ++i) {
		auto method_params = LeetCodeIO::split_array(params[i]);
		%s.at(%s[i])(method_params);
	}
`,
		systemDesignMethodListName,
		systemDesignMethodListName,
		systemDesignMethodMapName,
		systemDesignMethodListName,
	)
	return callCode
}

func (c cpp) generatePrintCode(q *leetcode.QuestionData) (printCode string) {
	if q.MetaData.SystemDesign {
		return fmt.Sprintf(
			"\tcout << \"\\n%s \" << LeetCodeIO::join_array(%s) << '\\n';\n",
			testCaseOutputMark,
			systemDesignOutputName,
		)
	}

	if q.MetaData.Return != nil && q.MetaData.Return.Type != "void" {
		printCode += "\t" + c.getPrintCodeForType(returnName, outputStreamName) + "\n"
	} else if q.MetaData.Output != nil {
		outputParamName := q.MetaData.Params[q.MetaData.Output.ParamIndex].Name
		printCode += "\t" + c.getPrintCodeForType(outputParamName, outputStreamName) + "\n"
	} else {
		printCode += fmt.Sprintf("\t%s << \"null\";\n", outputStreamName)
	}
	printCode += fmt.Sprintf("\tcout << \"\\n%s \" << %s.rdbuf() << '\\n';\n", testCaseOutputMark, outputStreamName)
	return printCode
}

func indentCode(code string) string {
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = "\t" + line
		}
	}
	return strings.Join(lines, "\n")
}

func (c cpp) generateTestContent(q *leetcode.QuestionData) (string, error) {
	const template = `int main() {
	ios_base::sync_with_stdio(false);
	try {
%s
	} catch (const LeetCodeIO::Error &error) {
		cerr << "LC_IO: " << error.what() << '\n';
		return 2;
	}
	return 0;
}`
	body := c.generateScanCode(q) + "\n" +
		c.generateInitCode(q) +
		c.generateCallCode(q)
	if !q.MetaData.SystemDesign {
		body += "\n\tstringstream " + outputStreamName + ";\n"
	}
	body += c.generatePrintCode(q)
	testContent := fmt.Sprintf(
		template,
		indentCode(strings.TrimSuffix(body, "\n")),
	)
	if q.MetaData.Manual {
		testContent = fmt.Sprintf("// %s\n%s", manualWarning, testContent)
	}
	return testContent, nil
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
	codeHeader := fmt.Sprintf(
		`#include <bits/stdc++.h>
#include "%s"
using namespace std;

`, cppUtils.HeaderName,
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

func (c cpp) RunLocalTest(q *leetcode.QuestionData, outDir string, targetCase string) (bool, error) {
	genResult, err := c.GeneratePaths(q)
	if err != nil {
		return false, fmt.Errorf("generate paths failed: %w", err)
	}
	genResult.SetOutDir(outDir)

	testFile := genResult.GetFile(TestFile).GetPath()
	if !utils.IsExist(testFile) {
		return false, fmt.Errorf("file %s not found", utils.RelToCwd(testFile))
	}
	execFile, err := getTempBinFile(q, c)
	if err != nil {
		return false, fmt.Errorf("generate temporary binary file path failed: %w", err)
	}

	cfg := config.Get()
	compilerFlags, _ := shlex.Split(cfg.Code.Cpp.CXXFLAGS)
	args := []string{cfg.Code.Cpp.CXX}
	args = append(args, compilerFlags...)
	args = append(args, "-I", outDir, "-o", execFile, testFile)

	err = buildTest(q, genResult, args)
	if err != nil {
		return false, fmt.Errorf("compilation failed: %w", err)
	}

	return runTest(q, genResult, []string{execFile}, targetCase)
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
