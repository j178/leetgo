package lang

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/log"
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

func addBinSection(result *GenerateResult) error {
	q := result.Question
	var mapping map[string]any
	cargoTomlPath := filepath.Join(result.OutDir, "Cargo.toml")
	data, err := os.ReadFile(cargoTomlPath)
	if err != nil {
		return err
	}
	err = toml.Unmarshal(data, &mapping)
	if err != nil {
		return err
	}
	bins, _ := mapping["bin"].([]any)
	exists := false
	for i, bin := range bins {
		if bin.(map[string]any)["name"] == q.TitleSlug {
			bins[i] = map[string]any{
				"name": q.TitleSlug,
				"path": filepath.Join(result.SubDir, "solution.rs"),
			}
			exists = true
			break
		}
	}
	if !exists {
		bins = append(
			bins, map[string]any{
				"name": q.TitleSlug,
				"path": filepath.Join(result.SubDir, "solution.rs"),
			},
		)
	}
	sort.Slice(
		bins, func(i, j int) bool {
			return bins[i].(map[string]any)["name"].(string) < bins[j].(map[string]any)["name"].(string)
		},
	)
	mapping["bin"] = bins
	data, err = toml.Marshal(mapping)
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
