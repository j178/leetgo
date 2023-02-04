package lang

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/j178/leetgo/leetcode"
)

type cpp struct {
	baseLang
}

//go:embed cpp/LC_DS.h
var structDefi string

//go:embed cpp/LC_IO.h
var helperFuncs string

var cppTypes = map[string]string{
	"void":     "void",
	"integer":  "int",
	"long":     "int64_t",
	"string":   "string",
	"double":   "double",
	"ListNode": "ListNode*",
	"TreeNode": "TreeNode*",
}

const (
	objectName            = "obj"
	returnName            = "res"
	inputFileStreamName   = "ifs"
	outputFileStreamName  = "ofs"
	systemDesignFuncName  = "sys_design_func"
	systemDesignFuncNames = "sys_design_funcs"
	cppTestFileTemplate   = `#include <bits/stdc++.h>
using namespace std;

// definitions
%s

#include "solution.h"

// helper funcs
%s

// main func
int main(int argc, char **argv) {
	if (argc != 3) {
		return 1;
	}

	ifstream ` + inputFileStreamName + `(argv[1]);
	if (!` + inputFileStreamName + `.is_open()) {
		return 1;
	}

	ofstream ` + outputFileStreamName + `(argv[2]);
	if (!` + outputFileStreamName + `.is_open()) {
		return 1;
	}

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

	` + inputFileStreamName + `.close();
	` + outputFileStreamName + `.close();
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
	if d == 0 && t == "string" {
		return fmt.Sprintf("%s >> quoted(%s);", ifs, n)
	} else {
		return fmt.Sprintf("%s >> %s;", ifs, n)
	}
}

func (c cpp) getPrintCodeForType(d int, t string, n string, ofs string) string {
	if d == 0 {
		if t == "string" {
			return fmt.Sprintf("%s << quoted(%s);", ofs, n)
		}
		if t == "double" {
			return fmt.Sprintf("{ char buf[320]; sprintf(buf, \"%%.5f\", %s); %s << string(buf); }", n, ofs)
		}
		return fmt.Sprintf("%s << %s;", ofs, n)
	} else {
		return fmt.Sprintf("%s << %s;", ofs, n)
	}
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
			"\t%s %s; %s >> %s;\n",
			c.getVectorTypeName(1, "string"),
			systemDesignFuncNames,
			inputFileStreamName,
			systemDesignFuncNames,
		)
	}

	var scanCode string
	for _, param := range q.MetaData.Params {
		dimCnt, cppType := c.getCppTypeName(param.Type)
		scanCode += fmt.Sprintf(
			"\t%s %s\n",
			c.getDeclCodeForType(dimCnt, cppType, param.Name),
			c.getScanCodeForType(dimCnt, cppType, param.Name, inputFileStreamName),
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
					c.getScanCodeForType(dimCnt, cppType, param.Name, inputFileStreamName),
					inputFileStreamName,
				)
			}
		} else {
			callCode += fmt.Sprintf("\t\t\t%s.ignore();\n", inputFileStreamName)
		}
	}

	if q.MetaData.SystemDesign {
		className := q.MetaData.ClassName
		callCode += fmt.Sprintf("\tauto hash_value_%s = hash<string>()(\"%s\");\n", className, className)
		for _, method := range q.MetaData.Methods {
			callCode += fmt.Sprintf("\tauto hash_value_%s = hash<string>()(\"%s\");\n", method.Name, method.Name)
		}
		callCode += fmt.Sprintf("\t%s.ignore(); %s << '[';\n", inputFileStreamName, outputFileStreamName)
		callCode += fmt.Sprintf("\tfor (auto &&%s : %s) {\n", systemDesignFuncName, systemDesignFuncNames)
		/* iterate thru all function calls */ {
			callCode += fmt.Sprintf(
				"\t\t%s.ignore(); auto hash_value = hash<string>()(%s);\n",
				inputFileStreamName,
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
				callCode += fmt.Sprintf("\t\t\t%s << \"null,\";\n\t\t}", outputFileStreamName)
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
						c.getPrintCodeForType(dimCnt, returnType, functionCall, outputFileStreamName),
						outputFileStreamName,
					)
				} else {
					callCode += fmt.Sprintf(
						"\t\t\t%s;\n\t\t\t%s << \"null,\";\n\t\t}",
						functionCall,
						outputFileStreamName,
					)
				}
			}
			callCode += fmt.Sprintf(
				" else {\n\t\t\treturn 1;\n\t\t}\n\t\t%s.ignore();\n",
				inputFileStreamName,
			)
		}
		callCode += fmt.Sprintf(
			"\t}\n\t%s.seekp(-1, ios_base::end); %s << ']';\n",
			outputFileStreamName,
			outputFileStreamName,
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

func (c cpp) generatePrintCode(q *leetcode.QuestionData) string {
	if q.MetaData.SystemDesign {
		return ""
	}
	dimCnt, cppType := c.getCppTypeName(q.MetaData.Return.Type)
	return fmt.Sprintf("\t%s\n", c.getPrintCodeForType(dimCnt, cppType, returnName, outputFileStreamName))
}

func (c cpp) generateTest(q *leetcode.QuestionData, testcases string) string {
	content := fmt.Sprintf(testFileHeader, c.lineComment)
	content += fmt.Sprintf(
		cppTestFileTemplate,
		structDefi,
		helperFuncs,
		c.generateScanCode(q),
		c.generateInitCode(q),
		c.generateCallCode(q),
		c.generatePrintCode(q),
	)
	return content
}

func (c cpp) getJudgeResult(dimCnt int, returnType string, expectedOutput string, actualOutput string) (eq bool) {
	// type specific judge
	//  * double - not needed at this time, as all double results are sanitized as "%.5f"
	//  * integer, long, string - raw string comparison is enough
	/*if returnType == "double" {
		eps := 1e-5
		if dimCnt > 0 {
			return false
		}
		a, _ := strconv.ParseFloat(expectedOutput, 32)
		b, _ := strconv.ParseFloat(actualOutput, 32)
		eq = math.Abs(a-b) < eps
	} else*/{
		eq = expectedOutput == actualOutput
	}
	return
}

func (c cpp) RunLocalTest(q *leetcode.QuestionData, dir string) (bool, error) {
	red := color.New(color.FgRed).Add(color.Bold)
	green := color.New(color.FgGreen).Add(color.Bold)
	magenta := color.New(color.FgMagenta).Add(color.Bold)
	yellow := color.New(color.FgYellow).Add(color.Bold)
	blue := color.New(color.FgBlue).Add(color.Bold)

	// get path
	filenameTmpl := getFilenameTemplate(q, c)
	baseFilename, err := q.GetFormattedFilename(c.slug, filenameTmpl)
	if err != nil {
		return false, err
	}
	testFile := filepath.Join("cpp/", baseFilename, "solution.cpp")
	execFile := filepath.Join("cpp/", baseFilename, "solution.exec")
	inputFile := filepath.Join("cpp/", baseFilename, "input.txt")
	outputFile := filepath.Join("cpp/", baseFilename, "output.txt")

	// compile
	cmd := exec.Command("g++", "-O2", "-std=c++17", testFile, "-o", execFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	compileStart := time.Now()
	err = cmd.Run()
	if err != nil {
		yellow.Print("[CE]")
		fmt.Println(" Compile Error! :(")
		return false, err
	} else {
		elapsed := time.Since(compileStart)
		fmt.Println("Compilation Finished in", elapsed)
	}

	// execute
	var argsNum int
	if q.MetaData.SystemDesign {
		argsNum = 2
	} else {
		argsNum = len(q.MetaData.Params)
	}
	cases := q.GetTestCases()
	outputs := q.ParseExampleOutputs()
	for i := 0; i < len(cases) && i/argsNum < len(outputs); i += argsNum {
		caseId := i / argsNum

		// write input file
		input := strings.Join(cases[i:i+argsNum], "\n")
		f, err := os.Create(inputFile)
		if err != nil {
			return false, err
		}
		_, err = f.WriteString(input)
		if err != nil {
			return false, err
		}
		err = f.Sync()
		if err != nil {
			return false, err
		}

		// execute test program
		timeLimit := 3000 * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), timeLimit)
		defer cancel()
		executeStart := time.Now()
		err = exec.CommandContext(ctx, execFile, inputFile, outputFile).Run()
		elapsed := time.Since(executeStart)
		if ctx.Err() != nil {
			blue.Print("[TLE]")
			fmt.Printf(" Case %d - Time Limit (%s) Exceeded! :(\n", caseId, timeLimit)
			continue
		}
		if err != nil {
			magenta.Print("[RE]")
			fmt.Printf(" Case %d - Runtime Error! :(\n", caseId)
			continue
		}

		// read test result
		filebuffer, err := os.ReadFile(outputFile)
		if err != nil {
			return false, err
		}
		expectedOutput := outputs[caseId]
		actualOutput := string(filebuffer)

		// judge & show result
		dimCnt, returnType := c.getCppTypeName(q.MetaData.Return.Type)
		if c.getJudgeResult(dimCnt, returnType, expectedOutput, actualOutput) {
			green.Print("[AC]")
			fmt.Printf(" Case %d - %s\n", caseId, elapsed)
			fmt.Printf(" - Input:      %s\n", strings.ReplaceAll(input, "\n", "↩ "))
			fmt.Printf(" - Output:     %s\n", actualOutput)
		} else {
			red.Print("[WA]")
			fmt.Printf(" Case %d - %s\n", caseId, elapsed)
			fmt.Printf(" - Input:      %s\n", strings.ReplaceAll(input, "\n", "↩ "))
			fmt.Printf(" - Output:     %s\n", actualOutput)
			fmt.Printf(" - Expected:   %s\n", expectedOutput)
		}
	}

	return true, nil
}

func (c cpp) Generate(q *leetcode.QuestionData) (*GenerateResult, error) {
	blocks := getBlocks(c)
	modifiers, err := getModifiers(c, goBuiltinModifiers)
	if err != nil {
		return nil, err
	}
	codeContent, err := c.generateContent(q, blocks, modifiers)
	if err != nil {
		return nil, err
	}

	testcaseStr := c.generateTestCases(q)
	testContent := c.generateTest(q, testcaseStr)

	filenameTmpl := getFilenameTemplate(q, c)
	baseFilename, err := q.GetFormattedFilename(c.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	codeFile := filepath.Join(baseFilename, "solution.h")
	testFile := filepath.Join(baseFilename, "solution.cpp")

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
		Lang:     c,
		Files:    files,
	}, nil
}

func (c cpp) GeneratePaths(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, c)
	baseFilename, err := q.GetFormattedFilename(c.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	codeFile := filepath.Join(baseFilename, "solution.h")
	testFile := filepath.Join(baseFilename, "solution.cpp")

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
		Lang:     c,
		Files:    files,
	}, nil
}
