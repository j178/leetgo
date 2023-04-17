package lang

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/constants"
	"github.com/pelletier/go-toml/v2"

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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = outDir
	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("cargo", "add", "serde", "serde_json", "anyhow", constants.RustTestUtilsCrate)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = outDir
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (r rust) RunLocalTest(q *leetcode.QuestionData, outDir string) (bool, error) {
	genResult, err := r.GeneratePaths(q)
	if err != nil {
		return false, fmt.Errorf("generate paths failed: %w", err)
	}
	genResult.SetOutDir(outDir)

	build := exec.Command("cargo", "build", "--bin", q.TitleSlug)
	build.Dir = outDir
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	log.Info("building", "cmd", build.String())
	err = build.Run()
	if err != nil {
		return false, fmt.Errorf("build failed: %w", err)
	}

	return runTest(q, genResult, []string{"cargo", "run", "--bin", q.TitleSlug}, outDir)
}

func convertToRustType(typeName string) string {
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
			return "Vec<" + convertToGoType(typeName[:len(typeName)-2]) + ">"
		}
	}
	return typeName
}

func toRustFuncName(name string) string {
	var sb strings.Builder
	for i, c := range name {
		if i > 0 && c >= 'A' && c <= 'Z' {
			sb.WriteRune('_')
		}
		sb.WriteRune(c)
	}
	return strings.ToLower(sb.String())
}

func (r rust) generateNormalTestCode(q *leetcode.QuestionData) (string, error) {
	const template = `fn main() -> Result<()> {
%s

	Ok(())
}`
	code := ""
	paramNames := make([]string, 0, len(q.MetaData.Params))
	for _, param := range q.MetaData.Params {
		code += fmt.Sprintf(
			"\tlet %s: %s = deserialize(read_line()?)?;\n",
			toRustFuncName(param.Name),
			convertToRustType(param.Type),
		)
		paramNames = append(paramNames, toRustFuncName(param.Name))
	}

	if q.MetaData.Return != nil && q.MetaData.Return.Type != "void" {
		code += fmt.Sprintf(
			"\tlet ans = Solution::%s(%s);\n",
			toRustFuncName(q.MetaData.Name),
			strings.Join(paramNames, ", "),
		)
	} else {
		code += fmt.Sprintf(
			"\tSolution::%s(%s);\n",
			toRustFuncName(q.MetaData.Name),
			strings.Join(paramNames, ", "),
		)
		ansName := paramNames[q.MetaData.Output.ParamIndex]
		code += fmt.Sprintf("\tlet ans = %s;\n", ansName)
	}
	code += fmt.Sprintf(
		"\tprintln!(\"%s {}\", serialize(ans)?);\n",
		testCaseOutputMark,
	)

	testContent := fmt.Sprintf(template, code)
	return testContent, nil
}

func (r rust) generateSystemDesignTestCode(q *leetcode.QuestionData) (string, error) {
	return "", nil
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
	codeHeader := fmt.Sprintf(
		`use anyhow::Result;

use %s::*;
`, constants.RustTestUtilsCrate,
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
	modifiers, err := getModifiers(r, goBuiltinModifiers)
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
