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
