package lang

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/j178/leetgo/leetcode"
	cppTestUtils "github.com/j178/leetgo/testutils/cpp"
	"github.com/j178/leetgo/utils"
)

type cpp struct {
	baseLang
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
	objectName            = "obj"
	returnName            = "res"
	inputStreamName       = "cin"
	outputStreamName      = "out_stream"
	outputMarker          = "output: "
	systemDesignFuncName  = "sys_design_func"
	systemDesignFuncNames = "sys_design_funcs"
	cppTestFileTemplate   = `#include "` + cppTestUtils.HeaderName + `"

#include <bits/stdc++.h>
using namespace std;

#include "solution.h"

// main func
int main() {
	// global init
	ios::sync_with_stdio(false);
	stringstream ` + outputStreamName + `;


	// scan input args
%s

	// initialize object
%s

	// call methods
%s

	// print result
%s

	// delete object
	delete ` + objectName + `;

	return 0;
}
`
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

func (c cpp) getScanCodeForType(d int, t string, n string, ifs string) string {
	if d == 0 {
		if t == "string" {
			return fmt.Sprintf("%s >> quoted(%s);", ifs, n)
		}
	}
	return fmt.Sprintf("%s >> %s;", ifs, n)
}

func (c cpp) getPrintCodeForType(d int, t string, n string, ofs string) string {
	if d == 0 {
		if t == "string" {
			return fmt.Sprintf("%s << quoted(%s);", ofs, n)
		}
		if t == "double" {
			return fmt.Sprintf("{ char buf[320]; sprintf(buf, \"%%.5f\", %s); %s << string(buf); }", n, ofs)
		}
	}
	return fmt.Sprintf("%s << %s;", ofs, n)
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
			c.getDeclCodeForType(1, "string", systemDesignFuncNames),
			c.getScanCodeForType(1, "string", systemDesignFuncNames, inputStreamName),
		)
	}

	var scanCode string
	for _, param := range q.MetaData.Params {
		dimCnt, cppType := c.getCppTypeName(param.Type)
		scanCode += fmt.Sprintf(
			"\t%s %s\n",
			c.getDeclCodeForType(dimCnt, cppType, param.Name),
			c.getScanCodeForType(dimCnt, cppType, param.Name, inputStreamName),
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

func (c cpp) generateCallCode(q *leetcode.QuestionData) string {
	var callCode string

	generateParamScanningCode := func(params []leetcode.MetaDataParam) {
		if len(params) > 0 {
			for _, param := range params {
				dimCnt, cppType := c.getCppTypeName(param.Type)
				callCode += fmt.Sprintf(
					"\t\t\t%s %s %s.ignore();\n",
					c.getDeclCodeForType(dimCnt, cppType, param.Name),
					c.getScanCodeForType(dimCnt, cppType, param.Name, inputStreamName),
					inputStreamName,
				)
			}
		} else {
			callCode += fmt.Sprintf("\t\t\t%s.ignore();\n", inputStreamName)
		}
	}

	if q.MetaData.SystemDesign {
		className := q.MetaData.ClassName
		callCode += fmt.Sprintf("\tauto hash_value_%s = hash<string>()(\"%s\");\n", className, className)
		for _, method := range q.MetaData.Methods {
			callCode += fmt.Sprintf("\tauto hash_value_%s = hash<string>()(\"%s\");\n", method.Name, method.Name)
		}
		callCode += fmt.Sprintf("\t%s.ignore(); %s << '[';\n", inputStreamName, outputStreamName)
		callCode += fmt.Sprintf("\tfor (auto &&%s : %s) {\n", systemDesignFuncName, systemDesignFuncNames)
		/* iterate thru all function calls */ {
			callCode += fmt.Sprintf(
				"\t\t%s.ignore(); auto hash_value = hash<string>()(%s);\n",
				inputStreamName,
				systemDesignFuncName,
			)
			/* operations in constructor function call */ {
				callCode += fmt.Sprintf("\t\tif (hash_value == hash_value_%s) {\n", className)
				generateParamScanningCode(q.MetaData.Constructor.Params)
				callCode += fmt.Sprintf(
					"\t\t\t%s = new %s(%s);\n",
					objectName,
					className,
					c.getParamString(q.MetaData.Constructor.Params),
				)
				callCode += fmt.Sprintf("\t\t\t%s << \"null,\";\n\t\t}", outputStreamName)
			}
			/* operations in member function calls */
			for _, method := range q.MetaData.Methods {
				callCode += fmt.Sprintf(" else if (hash_value == hash_value_%s) {\n", method.Name)
				generateParamScanningCode(method.Params)
				dimCnt, returnType := c.getCppTypeName(method.Return.Type)
				functionCall := fmt.Sprintf(
					"%s->%s(%s)",
					objectName,
					method.Name,
					c.getParamString(method.Params),
				)
				if returnType != "void" {
					callCode += fmt.Sprintf(
						"\t\t\t%s %s << ',';\n\t\t}",
						c.getPrintCodeForType(dimCnt, returnType, functionCall, outputStreamName),
						outputStreamName,
					)
				} else {
					callCode += fmt.Sprintf(
						"\t\t\t%s;\n\t\t\t%s << \"null,\";\n\t\t}",
						functionCall,
						outputStreamName,
					)
				}
			}
			callCode += fmt.Sprintf(
				" else {\n\t\t\treturn 1;\n\t\t}\n\t\t%s.ignore();\n",
				inputStreamName,
			)
		}
		callCode += fmt.Sprintf(
			"\t}\n\t%s.seekp(-1, ios_base::end); %s << ']';\n",
			outputStreamName,
			outputStreamName,
		)
	} else {
		callCode = fmt.Sprintf(
			"\tauto &&%s = %s->%s(%s);\n",
			returnName,
			objectName,
			q.MetaData.Name,
			c.getParamString(q.MetaData.Params),
		)
	}
	return callCode
}

func (c cpp) generatePrintCode(q *leetcode.QuestionData) (printCode string) {
	if !q.MetaData.SystemDesign {
		printCode += fmt.Sprintf("\t%s << %s;", outputStreamName, returnName)
	}
	printCode += fmt.Sprintf("\tcout << \"%s\" << %s.rdbuf();", outputMarker, outputStreamName)
	return
}

func (c cpp) generateTestContent(q *leetcode.QuestionData) string {
	return fmt.Sprintf(
		cppTestFileTemplate,
		c.generateScanCode(q),
		c.generateInitCode(q),
		c.generateCallCode(q),
		c.generatePrintCode(q),
	)
}

func (c cpp) generateTestFile(q *leetcode.QuestionData, filename string) (FileOutput, error) {
	content := c.generateTestContent(q)
	return FileOutput{
		Filename: filename,
		Content:  content,
		Type:     TestFile,
	}, nil
}

func (c cpp) RunLocalTest(q *leetcode.QuestionData, outDir string) (bool, error) {
	genResult, err := c.GeneratePaths(q)
	if err != nil {
		return false, fmt.Errorf("generate paths failed: %w", err)
	}
	genResult.SetOutDir(outDir)

	testFile := outDir + "/" + genResult.SubDir + "/solution.cpp"
	execFile := outDir + "/" + genResult.SubDir + "/solution.exec"

	cmd := exec.Command("g++", "-O2", "-std=c++17", "-I", outDir, "-o", execFile, testFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return false, fmt.Errorf("compilation failed: %w", err)
	}

	args := []string{execFile}
	return runTest(q, genResult, args, outDir)
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
	modifiers, err := getModifiers(c, goBuiltinModifiers)
	if err != nil {
		return nil, err
	}
	codeFile, err := c.generateCodeFile(q, "solution.h", blocks, modifiers, separateDescriptionFile)
	if err != nil {
		return nil, err
	}
	testFile, err := c.generateTestFile(q, "solution.cpp")
	if err != nil {
		return nil, err
	}
	testcaseFile, err := c.generateTestCasesFile(q, "testcases.txt")
	if err != nil {
		return nil, err
	}
	genResult.AddFile(codeFile)
	genResult.AddFile(testFile)
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
			Filename: "solution.h",
			Type:     CodeFile,
		},
	)
	genResult.AddFile(
		FileOutput{
			Filename: "solution.cpp",
			Type:     TestFile,
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

func (c cpp) Initialize(outDir string) error {
	headerPath := filepath.Join(outDir, cppTestUtils.HeaderName)
	if _, err := tryWrite(headerPath, cppTestUtils.HeaderContent); err != nil {
		return err
	}
	return nil
}

func (c cpp) HasInitialized(outDir string) (bool, error) {
	headerPath := filepath.Join(outDir, cppTestUtils.HeaderName)
	return utils.IsExist(headerPath), nil
}
