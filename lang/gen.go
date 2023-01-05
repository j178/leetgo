package lang

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
	"github.com/spf13/viper"
)

var (
	NotSupported   = errors.New("not supported")
	NotImplemented = errors.New("not implemented")
)

const (
	inputMark  = "input:"
	outputMark = "output:"
)

type GenerateResult struct {
	Generator Generator
	Files     []FileOutput
}

type FileOutput struct {
	Path    string
	Content string
	Written bool
	Type    FileType
}

type FileType int

const (
	CodeFile FileType = iota
	TestFile
	DocFile
	OtherFile
)

func (r *GenerateResult) GetCodeFile() *FileOutput {
	for _, f := range r.Files {
		if f.Type == CodeFile {
			return &f
		}
	}
	return nil
}

type Generator interface {
	Name() string
	ShortName() string
	Slug() string
	// Generate generates code files for the question.
	Generate(q *leetcode.QuestionData) (*GenerateResult, error)
	GeneratePaths(q *leetcode.QuestionData) (*GenerateResult, error)
	CodeBeginLine() string
	CodeEndLine() string
}

type Initializer interface {
	Initialized(dir string) (bool, error)
	Init(dir string) error
}

type LocalTester interface {
	Initializer
	RunTest(q *leetcode.QuestionData) error
}

type baseLang struct {
	name              string
	slug              string
	shortName         string
	extension         string
	lineComment       string
	blockCommentStart string
	blockCommentEnd   string
}

func (l baseLang) Name() string {
	return l.name
}

func (l baseLang) Slug() string {
	return l.slug
}

func (l baseLang) ShortName() string {
	return l.shortName
}

func (l baseLang) CodeBeginLine() string {
	return l.lineComment + " " + config.Get().Code.CodeBeginMark
}

func (l baseLang) CodeEndLine() string {
	return l.lineComment + " " + config.Get().Code.CodeEndMark
}

func (l baseLang) generateComments(q *leetcode.QuestionData) string {
	var content []string
	cfg := config.Get()
	now := time.Now().Format("2006/01/02 15:04")
	content = append(content, fmt.Sprintf("%s Created by %s at %s", l.lineComment, cfg.Author, now))
	content = append(content, fmt.Sprintf("%s %s", l.lineComment, q.Url()))
	if q.IsContest() {
		content = append(content, fmt.Sprintf("%s %s", l.lineComment, q.ContestUrl()))
	}
	content = append(content, "")
	content = append(content, l.blockCommentStart)
	content = append(content, fmt.Sprintf("%s.%s (%s)", q.QuestionFrontendId, q.GetTitle(), q.Difficulty))
	content = append(content, "")
	content = append(content, q.GetFormattedContent())
	content = append(content, l.blockCommentEnd)
	content = append(content, "")
	return strings.Join(content, "\n")
}

func (l baseLang) generateCode(q *leetcode.QuestionData, modifiers ...Modifier) string {
	code := q.GetCodeSnippet(l.Slug())
	for _, m := range modifiers {
		code = m(code, q)
	}
	return code
}

func needsDefinition(code string) bool {
	return strings.Contains(code, "Definition for")
}

func getFilenameTemplate(gen Generator) string {
	res := config.Get().Code.FilenameTemplate
	if res != "" {
		return res
	}
	res = viper.GetString("code." + gen.Slug() + ".filename_template")
	return res
}

func (l baseLang) generateTestCases(q *leetcode.QuestionData) string {
	cases := q.GetTestCases()
	outputs := q.ParseExampleOutputs()

	var caseAndOutputs []string
	for i, c := range cases {
		if i >= len(outputs) {
			break
		}
		caseAndOutputs = append(
			caseAndOutputs,
			fmt.Sprintf("%s\n%s\n%s\n%s", inputMark, c, outputMark, outputs[i]),
		)
	}
	return strings.Join(caseAndOutputs, "\n\n")
}

func (l baseLang) GeneratePaths(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(l)
	baseFilename, err := q.GetFormattedFilename(l.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}

	file := FileOutput{
		Path: baseFilename + l.extension,
		Type: CodeFile,
	}
	return &GenerateResult{
		Generator: l,
		Files:     []FileOutput{file},
	}, nil
}

func (l baseLang) Generate(q *leetcode.QuestionData) (*GenerateResult, error) {
	comment := l.generateComments(q)
	code := l.generateCode(q, addCodeMark(l))
	content := comment + "\n" + code + "\n"

	filenameTmpl := getFilenameTemplate(l)
	baseFilename, err := q.GetFormattedFilename(l.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}

	file := FileOutput{
		Path:    baseFilename + l.extension,
		Content: content,
		Type:    CodeFile,
	}
	return &GenerateResult{
		Generator: l,
		Files:     []FileOutput{file},
	}, nil
}

func GetGenerator(gen string) Generator {
	gen = strings.ToLower(gen)
	for _, l := range SupportedLangs {
		if l.Slug() == gen {
			return l
		}
	}
	for _, l := range SupportedLangs {
		if strings.HasPrefix(strings.ToLower(l.Name()), gen) {
			return l
		}
	}
	return nil
}

func Generate(q *leetcode.QuestionData) (*GenerateResult, error) {
	cfg := config.Get()
	gen := GetGenerator(cfg.Code.Lang)
	if gen == nil {
		return nil, fmt.Errorf("language %s is not supported yet, welcome to send a PR", cfg.Code.Lang)
	}

	err := q.Fulfill()
	if err != nil {
		return nil, fmt.Errorf("failed to get question data: %w", err)
	}

	codeSnippet := q.GetCodeSnippet(gen.Slug())
	if codeSnippet == "" {
		return nil, fmt.Errorf("no %s code snippet found for %s", cfg.Code.Lang, q.TitleSlug)
	}

	outDir := viper.GetString("code." + cfg.Code.Lang + ".out_dir")
	if outDir == "" {
		outDir = cfg.Code.Lang
	}
	outDir = filepath.Join(cfg.ProjectRoot(), outDir)

	err = utils.CreateIfNotExists(outDir, true)
	if err != nil {
		return nil, err
	}

	// Check and generate necessary library files.
	if t, ok := gen.(Initializer); ok {
		ok, err := t.Initialized(outDir)
		if err == nil && !ok {
			err = t.Init(outDir)
			if err != nil {
				return nil, err
			}
		}
		if err != nil {
			hclog.L().Error(
				"check library failed, skip library generation",
				"lang", gen.Slug(),
				"err", err,
			)
		}
	}

	// Generate files
	result, err := gen.Generate(q)
	if err != nil {
		return nil, err
	}

	// Write files
	for i := range result.Files {
		path := filepath.Join(outDir, result.Files[i].Path)
		result.Files[i].Path = path
		written, err := tryWrite(path, result.Files[i].Content)
		if err != nil {
			hclog.L().Error("failed to write file", "path", path, "err", err)
			continue
		}
		result.Files[i].Written = written
	}

	// Update last generated state
	state := config.LoadState()
	state.LastQuestion = config.LastQuestion{
		Slug:       q.TitleSlug,
		FrontendID: q.QuestionFrontendId,
		Gen:        gen.Slug(),
	}
	config.SaveState(state)

	return result, nil
}

func tryWrite(file string, content string) (bool, error) {
	write := true
	relPath := utils.RelToCwd(file)
	if utils.IsExist(file) {
		if !viper.GetBool("yes") {
			prompt := &survey.Confirm{Message: fmt.Sprintf("File \"%s\" already exists, overwrite?", relPath)}
			err := survey.AskOne(prompt, &write)
			if err != nil {
				return false, err
			}
		}
	}
	if !write {
		return false, nil
	}

	err := utils.CreateIfNotExists(file, false)
	if err != nil {
		return false, err
	}
	err = os.WriteFile(file, []byte(content), 0644)
	if err != nil {
		return false, err
	}
	hclog.L().Info("generated", "file", relPath)
	return true, nil
}

// GenerateDryRun runs generate process but does not generate real content.
func GenerateDryRun(q *leetcode.QuestionData) (*GenerateResult, error) {
	cfg := config.Get()
	gen := GetGenerator(cfg.Code.Lang)
	if gen == nil {
		return nil, fmt.Errorf("language %s is not supported", cfg.Code.Lang)
	}

	outDir := viper.GetString("code." + cfg.Code.Lang + ".out_dir")
	if outDir == "" {
		outDir = cfg.Code.Lang
	}
	outDir = filepath.Join(cfg.ProjectRoot(), outDir)

	result, err := gen.GeneratePaths(q)
	if err != nil {
		return nil, err
	}

	for i := range result.Files {
		path := filepath.Join(outDir, result.Files[i].Path)
		result.Files[i].Path = path
	}

	return result, nil
}

func GetSolutionCode(q *leetcode.QuestionData) (string, error) {
	result, err := GenerateDryRun(q)
	if err != nil {
		return "", err
	}
	codeFile := result.GetCodeFile()
	if codeFile == nil {
		return "", fmt.Errorf("no code file generated")
	}
	if !utils.IsExist(codeFile.Path) {
		return "", fmt.Errorf("code file %s does not exist", codeFile.Path)
	}
	code, err := os.ReadFile(codeFile.Path)
	if err != nil {
		return "", err
	}

	gen := result.Generator
	codeLines := strings.Split(string(code), "\n")
	var codeLinesToKeep []string
	inCode := false
	for _, line := range codeLines {
		if !inCode && strings.Contains(line, gen.CodeBeginLine()) {
			inCode = true
			continue
		}
		if inCode && strings.Contains(line, gen.CodeEndLine()) {
			break
		}
		if inCode {
			codeLinesToKeep = append(codeLinesToKeep, line)
		}
	}

	if len(codeLinesToKeep) == 0 {
		return "", fmt.Errorf("no code found in %s", codeFile)
	}

	return strings.Join(codeLinesToKeep, "\n"), nil
}
