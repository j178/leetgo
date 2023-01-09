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
	NotSupported = errors.New("not supported")
)

const (
	testCaseInputMark  = "input:"
	testCaseOutputMark = "output:"
)

type GenerateResult struct {
	Lang  Lang
	Files []FileOutput
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

func (r *GenerateResult) PrependPath(dir string) {
	for i, f := range r.Files {
		r.Files[i].Path = filepath.Join(dir, f.Path)
	}
}

type Lang interface {
	Name() string
	ShortName() string
	Slug() string
	LineComment() string
	// Generate generates code files for the question.
	Generate(q *leetcode.QuestionData) (*GenerateResult, error)
	GeneratePaths(q *leetcode.QuestionData) (*GenerateResult, error)
}

type NeedInitialization interface {
	HasInitialized(dir string) (bool, error)
	Initialize(dir string) error
}

type LocalTestable interface {
	RunLocalTest(q *leetcode.QuestionData, dir string) error
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

func (l baseLang) LineComment() string {
	return l.lineComment
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
			fmt.Sprintf("%s\n%s\n%s\n%s", testCaseInputMark, c, testCaseOutputMark, outputs[i]),
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
		Lang:  l,
		Files: []FileOutput{file},
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
		Lang:  l,
		Files: []FileOutput{file},
	}, nil
}

func GetGenerator(lang string) Lang {
	lang = strings.ToLower(lang)
	for _, l := range SupportedLangs {
		if l.Slug() == lang {
			return l
		}
	}
	for _, l := range SupportedLangs {
		if l.ShortName() == lang {
			return l
		}
	}
	for _, l := range SupportedLangs {
		if strings.HasPrefix(strings.ToLower(l.Name()), lang) {
			return l
		}
	}
	return nil
}

func getCodeConfig(lang Lang, key string) string {
	ans := viper.GetString("code." + lang.Slug() + "." + key)
	if ans != "" {
		return ans
	}
	return viper.GetString("code." + lang.ShortName() + "." + key)
}

func needsDefinition(code string) bool {
	return strings.Contains(code, "Definition for")
}

func getFilenameTemplate(gen Lang) string {
	ans := getCodeConfig(gen, "filename_template")
	if ans != "" {
		return ans
	}
	return config.Get().Code.FilenameTemplate
}

func getOutDir(lang Lang) string {
	cfg := config.Get()
	outDir := getCodeConfig(lang, "out_dir")
	if outDir == "" {
		outDir = lang.Slug()
	}
	outDir = filepath.Join(cfg.ProjectRoot(), outDir)
	return outDir
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

	outDir := getOutDir(gen)
	err = utils.CreateIfNotExists(outDir, true)
	if err != nil {
		return nil, err
	}

	// Check and generate necessary library files.
	if t, ok := gen.(NeedInitialization); ok {
		ok, err := t.HasInitialized(outDir)
		if err == nil && !ok {
			err = t.Initialize(outDir)
			if err != nil {
				return nil, err
			}
		}
		if err != nil {
			hclog.L().Error(
				"check initialization failed, skip initialization",
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
	result.PrependPath(outDir)
	// Write files
	for i, file := range result.Files {
		written, err := tryWrite(file.Path, file.Content)
		if err != nil {
			hclog.L().Error("failed to write file", "path", file.Path, "err", err)
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
	err = os.WriteFile(file, utils.StringToBytes(content), 0644)
	if err != nil {
		return false, err
	}
	hclog.L().Info("generated", "file", relPath)
	return true, nil
}

// GeneratePathsOnly runs generate process but does not generate real content.
func GeneratePathsOnly(q *leetcode.QuestionData) (*GenerateResult, error) {
	cfg := config.Get()
	gen := GetGenerator(cfg.Code.Lang)
	if gen == nil {
		return nil, fmt.Errorf("language %s is not supported yet", cfg.Code.Lang)
	}

	result, err := gen.GeneratePaths(q)
	if err != nil {
		return nil, err
	}

	outDir := getOutDir(gen)
	result.PrependPath(outDir)
	return result, nil
}

func GetSolutionCode(q *leetcode.QuestionData) (string, error) {
	result, err := GeneratePathsOnly(q)
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

	cfg := config.Get()
	codeLines := strings.Split(string(code), "\n")
	var codeLinesToKeep []string
	inCode := false
	for _, line := range codeLines {
		if !inCode && strings.Contains(line, cfg.Code.CodeBeginMark) {
			inCode = true
			continue
		}
		if inCode && strings.Contains(line, cfg.Code.CodeEndMark) {
			break
		}
		if inCode {
			codeLinesToKeep = append(codeLinesToKeep, line)
		}
	}

	if len(codeLinesToKeep) == 0 {
		return "", fmt.Errorf("no code found in %s", codeFile.Path)
	}

	return strings.Join(codeLinesToKeep, "\n"), nil
}

func RunLocalTest(q *leetcode.QuestionData) error {
	cfg := config.Get()
	gen := GetGenerator(cfg.Code.Lang)
	if gen == nil {
		return fmt.Errorf("language %s is not supported", cfg.Code.Lang)
	}

	tester, ok := gen.(LocalTestable)
	if !ok {
		return fmt.Errorf("language %s does not support local test", gen.Slug())
	}

	outDir := getOutDir(gen)
	if !utils.IsExist(outDir) {
		return fmt.Errorf("no code generated for %s in language %s", q.TitleSlug, gen.Slug())
	}

	return tester.RunLocalTest(q, outDir)
}
