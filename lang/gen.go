package lang

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/charmbracelet/log"
	"github.com/spf13/viper"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

const (
	testCaseInputMark  = "input:"
	testCaseOutputMark = "output:"
)

type GenerateResult struct {
	Question *leetcode.QuestionData
	Lang     Lang
	Files    []FileOutput
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
	RunLocalTest(q *leetcode.QuestionData, dir string) (bool, error)
}

func GetGenerator(lang string) (Lang, error) {
	lang = strings.ToLower(lang)
	for _, l := range SupportedLangs {
		if l.Slug() == lang {
			return l, nil
		}
	}
	for _, l := range SupportedLangs {
		if l.ShortName() == lang {
			return l, nil
		}
	}
	for _, l := range SupportedLangs {
		if strings.HasPrefix(strings.ToLower(l.Name()), lang) {
			return l, nil
		}
	}
	return nil, fmt.Errorf("language %s is not supported yet, welcome to send a PR", lang)
}

func generate(q *leetcode.QuestionData) (Lang, *GenerateResult, error) {
	cfg := config.Get()
	gen, err := GetGenerator(cfg.Code.Lang)
	if err != nil {
		return nil, nil, err
	}

	err = q.Fulfill()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get question data: %w", err)
	}

	codeSnippet := q.GetCodeSnippet(gen.Slug())
	if codeSnippet == "" {
		return nil, nil, fmt.Errorf(`question "%s" doesn't support using "%s"`, cfg.Code.Lang, q.TitleSlug)
	}

	outDir := getOutDir(q, gen)
	err = utils.CreateIfNotExists(outDir, true)
	if err != nil {
		return nil, nil, err
	}

	// Check and generate necessary library files.
	if t, ok := gen.(NeedInitialization); ok {
		ok, err := t.HasInitialized(outDir)
		if err == nil && !ok {
			err = t.Initialize(outDir)
			if err != nil {
				return nil, nil, err
			}
		}
		if err != nil {
			log.Error(
				"check initialization failed, skip initialization",
				"lang", gen.Slug(),
				"err", err,
			)
		}
	}

	// Generate files
	result, err := gen.Generate(q)
	if err != nil {
		return nil, nil, err
	}
	result.PrependPath(outDir)
	// Write files
	for i, file := range result.Files {
		written, err := tryWrite(file.Path, file.Content)
		if err != nil {
			log.Error("failed to write file", "path", file.Path, "err", err)
			continue
		}
		result.Files[i].Written = written
	}
	return gen, result, nil
}

func Generate(q *leetcode.QuestionData) (*GenerateResult, error) {
	gen, result, err := generate(q)
	if err != nil {
		return nil, err
	}

	state := config.LoadState()
	state.LastQuestion = config.LastQuestion{
		Slug:       q.TitleSlug,
		FrontendID: q.QuestionFrontendId,
		Gen:        gen.Slug(),
	}
	config.SaveState(state)

	return result, nil
}

func GenerateContest(ct *leetcode.Contest) ([]*GenerateResult, error) {
	qs, err := ct.GetAllQuestions()
	if err != nil {
		return nil, err
	}

	var results []*GenerateResult
	for _, q := range qs {
		_, result, err := generate(q)
		if err != nil {
			log.Error("failed to generate", "question", q.TitleSlug, "err", err)
			continue
		}
		results = append(results, result)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no question generated")
	}

	state := config.LoadState()
	state.LastContest = ct.TitleSlug
	config.SaveState(state)

	return results, nil
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
	err = os.WriteFile(file, utils.StringToBytes(content), 0o644)
	if err != nil {
		return false, err
	}
	log.Info("generated", "file", relPath)
	return true, nil
}

// GeneratePathsOnly runs generate process but does not generate real content.
func GeneratePathsOnly(q *leetcode.QuestionData) (*GenerateResult, error) {
	cfg := config.Get()
	gen, err := GetGenerator(cfg.Code.Lang)
	if err != nil {
		return nil, err
	}

	result, err := gen.GeneratePaths(q)
	if err != nil {
		return nil, err
	}

	outDir := getOutDir(q, gen)
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

	codeLines := strings.Split(string(code), "\n")
	var codeLinesToKeep []string
	inCode := false
	for _, line := range codeLines {
		if !inCode && strings.Contains(line, config.CodeBeginMarker) {
			inCode = true
			continue
		}
		if inCode && strings.Contains(line, config.CodeEndMarker) {
			break
		}
		if inCode {
			codeLinesToKeep = append(codeLinesToKeep, line)
		}
	}

	nonEmptyLines := 0
	for _, line := range codeLinesToKeep {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}
	if nonEmptyLines == 0 {
		return "", fmt.Errorf("no code found in %s", codeFile.Path)
	}

	return strings.Join(codeLinesToKeep, "\n"), nil
}

func UpdateSolutionCode(q *leetcode.QuestionData, newCode string) error {
	result, err := GeneratePathsOnly(q)
	if err != nil {
		return err
	}
	codeFile := result.GetCodeFile()
	if codeFile == nil {
		return fmt.Errorf("no code file generated")
	}
	if !utils.IsExist(codeFile.Path) {
		return fmt.Errorf("code file %s does not exist", codeFile.Path)
	}
	code, err := os.ReadFile(codeFile.Path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(code), "\n")
	var newLines []string
	skip := false
	for _, line := range lines {
		if strings.Contains(line, config.CodeBeginMarker) {
			newLines = append(newLines, line)
			newLines = append(newLines, newCode)
			skip = true
		} else if strings.Contains(line, config.CodeEndMarker) {
			newLines = append(newLines, line)
			skip = false
		} else if !skip {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n")
	err = os.WriteFile(codeFile.Path, []byte(newContent), 0o644)
	if err != nil {
		return err
	}
	log.Info("updated", "file", utils.RelToCwd(codeFile.Path))
	return nil
}

func RunLocalTest(q *leetcode.QuestionData) (bool, error) {
	cfg := config.Get()
	gen, err := GetGenerator(cfg.Code.Lang)
	if err != nil {
		return false, err
	}

	tester, ok := gen.(LocalTestable)
	if !ok {
		return false, fmt.Errorf("language %s does not support local test", gen.Slug())
	}

	outDir := getOutDir(q, gen)
	if !utils.IsExist(outDir) {
		return false, fmt.Errorf("no code generated for %s in language %s", q.TitleSlug, gen.Slug())
	}

	return tester.RunLocalTest(q, outDir)
}
