package lang

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/pelletier/go-toml/v2"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/constants"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

type rust struct {
	baseLang
}

func (r rust) HasInitialized(outDir string) (bool, error) {
	if !utils.IsExist(filepath.Join(outDir, "Cargo.toml")) {
		return false, nil
	}
	return true, nil
}

func (r rust) Initialize(outDir string) error {
	const packageName = "leetcode-solutions"
	cmd := exec.Command("cargo", "init", "--bin", "--name", packageName, outDir)
	log.Info("cargo init", "cmd", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = outDir
	err := cmd.Run()
	if err != nil {
		return err
	}
	cmd = exec.Command("cargo", "add", "serde", "serde_json", "anyhow", constants.RustTestUtilsCrate)
	log.Info("cargo add", "cmd", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = outDir
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (r rust) RunLocalTest(q *leetcode.QuestionData, outDir string, targetCase string) (bool, error) {
	genResult, err := r.GeneratePaths(q)
	if err != nil {
		return false, fmt.Errorf("generate paths failed: %w", err)
	}
	genResult.SetOutDir(outDir)

	testFile := genResult.GetFile(TestFile).GetPath()
	if !utils.IsExist(testFile) {
		return false, fmt.Errorf("file %s not found", utils.RelToCwd(testFile))
	}

	args := []string{"cargo", "build", "--quiet", "--bin", q.TitleSlug}
	err = buildTest(q, genResult, args)
	if err != nil {
		return false, fmt.Errorf("build failed: %w", err)
	}

	return runTest(q, genResult, []string{"cargo", "run", "--quiet", "--bin", q.TitleSlug}, targetCase)
}

func toRustType(typeName string) string {
	switch typeName {
	case "integer":
		return "i32"
	case "string":
		return "String"
	case "long":
		return "i64"
	case "double":
		return "f64"
	case "boolean":
		return "bool"
	case "character":
		return "char"
	case "void":
		return "!"
	case "TreeNode":
		return "BinaryTree"
	case "ListNode":
		return "LinkedList"
	default:
		if strings.HasSuffix(typeName, "[]") {
			return "Vec<" + toRustType(typeName[:len(typeName)-2]) + ">"
		}
	}
	return typeName
}

func toRustVarName(name string) string {
	return utils.CamelToSnake(name)
}

func formatCallArgs(argTypes, args []string) string {
	if len(args) == 0 {
		return ""
	}
	res := make([]string, 0, len(args))
	for i, arg := range args {
		if argTypes[i] == "BinaryTree" || argTypes[i] == "LinkedList" {
			res = append(res, arg+".into()")
		} else {
			res = append(res, arg)
		}
	}
	return strings.Join(res, ", ")
}

func (r rust) generateNormalTestCode(q *leetcode.QuestionData) (string, error) {
	const template = `fn main() -> Result<()> {
%s
	println!("\n%s {}", serialize(ans)?);
	Ok(())
}`
	code := ""
	paramTypes := make([]string, 0, len(q.MetaData.Params))
	paramNames := make([]string, 0, len(q.MetaData.Params))
	for _, param := range q.MetaData.Params {
		varName := toRustVarName(param.Name)
		varType := toRustType(param.Type)
		code += fmt.Sprintf(
			"\tlet %s: %s = deserialize(&read_line()?)?;\n",
			varName,
			varType,
		)
		if !param.HelperParam {
			paramNames = append(paramNames, varName)
			paramTypes = append(paramTypes, varType)
		}
	}

	if q.MetaData.Return != nil && q.MetaData.Return.Type != "void" {
		code += fmt.Sprintf(
			"\tlet ans: %s = Solution::%s(%s).into();\n",
			toRustType(q.MetaData.Return.Type),
			toRustVarName(q.MetaData.Name),
			formatCallArgs(paramTypes, paramNames),
		)
	} else {
		// TODO: input param should be mut ref
		code += fmt.Sprintf(
			"\tSolution::%s(%s);\n",
			toRustVarName(q.MetaData.Name),
			formatCallArgs(paramTypes, paramNames),
		)
		ansName := paramNames[q.MetaData.Output.ParamIndex]
		ansType := paramTypes[q.MetaData.Output.ParamIndex]
		code += fmt.Sprintf(
			"\tlet ans: %s = %s.into();\n",
			toRustType(ansType),
			toRustVarName(ansName),
		)
	}

	testContent := fmt.Sprintf(template, code, testCaseOutputMark)
	return testContent, nil
}

func (r rust) generateSystemDesignTestCode(q *leetcode.QuestionData) (string, error) {
	const template = `fn main() -> Result<()> {
	let ops: Vec<String> = deserialize(&read_line()?)?;
	let params = split_array(&read_line()?)?;
	let mut output = Vec::with_capacity(ops.len());
	output.push("null".to_string());

%s

	for i in 1..ops.len() {
		match ops[i].as_str() {
%s
			_ => panic!("unknown op"),
		}
	}

	println!("\n%s {}", join_array(output));
	Ok(())
}
`
	var prepareCode string
	paramNames := make([]string, 0, len(q.MetaData.Constructor.Params))
	paramTypes := make([]string, 0, len(q.MetaData.Constructor.Params))
	if len(q.MetaData.Constructor.Params) > 0 {
		prepareCode += "\tlet constructor_params = split_array(&params[0])?;\n"
		for i, param := range q.MetaData.Constructor.Params {
			varName := toRustVarName(param.Name)
			varType := toRustType(param.Type)
			prepareCode += fmt.Sprintf(
				"\tlet %s: %s = deserialize(&constructor_params[%d])?;\n",
				varName,
				varType,
				i,
			)
			paramNames = append(paramNames, varName)
			paramTypes = append(paramTypes, varType)
		}
	}
	prepareCode += fmt.Sprintf(
		"\t#[allow(unused_mut)]\n\tlet mut obj = %s::new(%s);",
		q.MetaData.ClassName,
		formatCallArgs(paramTypes, paramNames),
	)

	callCode := ""
	for _, method := range q.MetaData.Methods {
		methodCall := fmt.Sprintf("\t\t\t\"%s\" => {\n", method.Name)
		if len(method.Params) > 0 {
			methodCall += "\t\t\t\tlet method_params = split_array(&params[i])?;\n"
		}
		methodParamNames := make([]string, 0, len(method.Params))
		methodParamTypes := make([]string, 0, len(method.Params))
		for i, param := range method.Params {
			varName := toRustVarName(param.Name)
			varType := toRustType(param.Type)
			methodCall += fmt.Sprintf(
				"\t\t\t\tlet %s: %s = deserialize(&method_params[%d])?;\n",
				varName,
				varType,
				i,
			)
			methodParamNames = append(methodParamNames, varName)
			methodParamTypes = append(methodParamTypes, varType)
		}

		if method.Return.Type != "" && method.Return.Type != "void" {
			methodCall += fmt.Sprintf(
				"\t\t\t\tlet ans: %s = obj.%s(%s).into();\n\t\t\t\toutput.push(serialize(ans)?);\n",
				toRustType(method.Return.Type),
				toRustVarName(method.Name),
				formatCallArgs(methodParamTypes, methodParamNames),
			)
		} else {
			methodCall += fmt.Sprintf(
				"\t\t\t\tobj.%s(%s);\n",
				toRustVarName(method.Name),
				formatCallArgs(methodParamTypes, methodParamNames),
			)
			methodCall += "\t\t\t\toutput.push(\"null\".to_string());\n"
		}
		methodCall += "\t\t\t}\n"
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

func (r rust) generateTestContent(q *leetcode.QuestionData) (string, error) {
	if q.MetaData.SystemDesign {
		return r.generateSystemDesignTestCode(q)
	}
	return r.generateNormalTestCode(q)
}

func (r rust) generateCodeFile(
	q *leetcode.QuestionData,
	filename string,
	blocks []config.Block,
	modifiers []ModifierFunc,
	separateDescriptionFile bool,
) (
	FileOutput,
	error,
) {
	var emptySolution string
	if !q.MetaData.SystemDesign {
		emptySolution = "\nstruct Solution;\n"
	}
	codeHeader := fmt.Sprintf(
		`use anyhow::Result;
use %s::*;
%s
`, constants.RustTestUtilsCrate,
		emptySolution,
	)
	testContent, err := r.generateTestContent(q)
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
	content, err := r.generateCodeContent(
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

func (r rust) GeneratePaths(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, r)
	baseFilename, err := q.GetFormattedFilename(r.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	genResult := &GenerateResult{
		SubDir:   filepath.Join("src", baseFilename),
		Question: q,
		Lang:     r,
	}
	genResult.AddFile(
		FileOutput{
			Filename: "solution.rs",
			Type:     CodeFile | TestFile,
		},
	)
	genResult.AddFile(
		FileOutput{
			Filename: "testcases.txt",
			Type:     TestCasesFile,
		},
	)
	if separateDescriptionFile(r) {
		genResult.AddFile(
			FileOutput{
				Filename: "question.md",
				Type:     DocFile,
			},
		)
	}
	return genResult, nil
}

type CargoToml struct {
	CargoFeatures     []string         `toml:"cargo-features,omitempty"`
	Package           map[string]any   `toml:"package,omitempty"`
	Lib               map[string]any   `toml:"lib,omitempty"`
	Bin               []map[string]any `toml:"bin,omitempty"`
	Example           []map[string]any `toml:"example,omitempty"`
	Test              []map[string]any `toml:"test,omitempty"`
	Bench             []map[string]any `toml:"bench,omitempty"`
	Dependencies      map[string]any   `toml:"dependencies"`
	DevDependencies   map[string]any   `toml:"dev-dependencies,omitempty"`
	BuildDependencies map[string]any   `toml:"build-dependencies,omitempty"`
	Target            map[string]any   `toml:"target,omitempty"`
	Badge             map[string]any   `toml:"badge,omitempty"`
	Features          map[string]any   `toml:"features,omitempty"`
	Patch             map[string]any   `toml:"patch,omitempty"`
	Replace           map[string]any   `toml:"replace,omitempty"`
	Profile           map[string]any   `toml:"profile,omitempty"`
	Workspace         map[string]any   `toml:"workspace,omitempty"`
}

func addBinSection(result *GenerateResult) error {
	q := result.Question
	cargoTomlPath := filepath.Join(result.OutDir, "Cargo.toml")
	data, err := os.ReadFile(cargoTomlPath)
	if err != nil {
		return err
	}

	var cargo CargoToml
	err = toml.Unmarshal(data, &cargo)
	if err != nil {
		return err
	}

	exists := false
	for i, bin := range cargo.Bin {
		if bin["name"].(string) == q.TitleSlug {
			cargo.Bin[i] = map[string]any{
				"name": q.TitleSlug,
				"path": filepath.Join(result.SubDir, "solution.rs"),
			}
			exists = true
			break
		}
	}
	if !exists {
		cargo.Bin = append(
			cargo.Bin, map[string]any{
				"name": q.TitleSlug,
				"path": filepath.Join(result.SubDir, "solution.rs"),
			},
		)
	}
	sort.Slice(
		cargo.Bin, func(i, j int) bool {
			return cargo.Bin[i]["path"].(string) < cargo.Bin[j]["path"].(string)
		},
	)

	data, err = toml.Marshal(cargo)
	if err != nil {
		return err
	}

	return os.WriteFile(cargoTomlPath, data, 0o644)
}

func (r rust) Generate(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, r)
	baseFilename, err := q.GetFormattedFilename(r.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	genResult := &GenerateResult{
		Question: q,
		Lang:     r,
		SubDir:   filepath.Join("src", baseFilename),
	}

	separateDescriptionFile := separateDescriptionFile(r)
	blocks := getBlocks(r)
	modifiers, err := getModifiers(r, builtinModifiers)
	if err != nil {
		return nil, err
	}
	codeFile, err := r.generateCodeFile(q, "solution.rs", blocks, modifiers, separateDescriptionFile)
	if err != nil {
		return nil, err
	}
	testcaseFile, err := r.generateTestCasesFile(q, "testcases.txt")
	if err != nil {
		return nil, err
	}
	genResult.AddFile(codeFile)
	genResult.AddFile(testcaseFile)

	if separateDescriptionFile {
		docFile, err := r.generateDescriptionFile(q, "question.md")
		if err != nil {
			return nil, err
		}
		genResult.AddFile(docFile)
	}

	// Add new [[bin]] section to Cargo.toml
	genResult.ResultHooks = append(genResult.ResultHooks, addBinSection)

	return genResult, nil
}
