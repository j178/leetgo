package lang

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
	"github.com/spf13/viper"
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

func (l baseLang) generateContent(q *leetcode.QuestionData, blocks []config.Block, modifiers []ModifierFunc) (
	string,
	error,
) {
	code := q.GetCodeSnippet(l.Slug())
	tmpl := template.New("root")
	tmpl.Funcs(
		template.FuncMap{
			"runModifiers": func(code string) string {
				for _, m := range modifiers {
					code = m(code, q)
				}
				return code
			},
		},
	)
	_, err := tmpl.Parse(contentTemplate)
	if err != nil {
		return "", err
	}

	for _, block := range blocks {
		if _, ok := validBlocks[block.Name]; !ok {
			return "", fmt.Errorf("invalid block name: %s", block.Name)
		}
		_, err := tmpl.New(block.Name).Parse(block.Template)
		if err != nil {
			return "", err
		}
	}

	cfg := config.Get()
	data := &contentData{
		Question:          q,
		Author:            cfg.Author,
		Time:              time.Now().Format("2006/01/02 15:04"),
		LineComment:       l.lineComment,
		BlockCommentStart: l.blockCommentStart,
		BlockCommentEnd:   l.blockCommentEnd,
		CodeBeginMarker:   config.CodeBeginMarker,
		CodeEndMarker:     config.CodeEndMarker,
		Code:              code,
		NeedsDefinition:   needsDefinition(code),
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (l baseLang) generateTestCases(q *leetcode.QuestionData) string {
	cases := q.GetTestCases()
	outputs := q.ParseExampleOutputs()
	argsNum := 0
	if q.MetaData.SystemDesign {
		argsNum = 2
	} else {
		argsNum = len(q.MetaData.Params)
	}

	// Assume all questions output are single.
	var caseAndOutputs []string
	for i := 0; i < len(cases) && i/argsNum < len(outputs); i += argsNum {
		input := strings.Join(cases[i:i+argsNum], "\n")
		caseAndOutputs = append(
			caseAndOutputs,
			fmt.Sprintf("%s\n%s\n%s\n%s", testCaseInputMark, input, testCaseOutputMark, outputs[i/argsNum]),
		)
	}
	return strings.Join(caseAndOutputs, "\n\n")
}

func (l baseLang) GeneratePaths(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, l)
	baseFilename, err := q.GetFormattedFilename(l.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}

	file := FileOutput{
		Path: baseFilename + l.extension,
		Type: CodeFile,
	}
	return &GenerateResult{
		Question: q,
		Lang:     l,
		Files:    []FileOutput{file},
	}, nil
}

func (l baseLang) Generate(q *leetcode.QuestionData) (*GenerateResult, error) {
	blocks := getBlocks(l)
	modifiers, err := getModifiers(l, builtinModifiers)
	if err != nil {
		return nil, err
	}
	content, err := l.generateContent(q, blocks, modifiers)
	if err != nil {
		return nil, err
	}

	filenameTmpl := getFilenameTemplate(q, l)
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
		Question: q,
		Lang:     l,
		Files:    []FileOutput{file},
	}, nil
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

func getCodeConfig(lang Lang, key string) string {
	ans := viper.GetString("code." + lang.Slug() + "." + key)
	if ans != "" {
		return ans
	}
	return viper.GetString("code." + lang.ShortName() + "." + key)
}

func getFilenameTemplate(q *leetcode.QuestionData, gen Lang) string {
	if q.IsContest() {
		return config.Get().Contest.FilenameTemplate
	}
	ans := getCodeConfig(gen, "filename_template")
	if ans != "" {
		return ans
	}
	return config.Get().Code.FilenameTemplate
}

func getOutDir(q *leetcode.QuestionData, lang Lang) string {
	if q.IsContest() {
		return config.Get().Contest.OutDir
	}
	cfg := config.Get()
	outDir := getCodeConfig(lang, "out_dir")
	// If outDir is not set, use the language slug as the outDir.
	if outDir == "" {
		outDir = lang.Slug()
	}
	outDir = filepath.Join(cfg.ProjectRoot(), outDir)
	return outDir
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
		return nil, nil, err
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
			hclog.L().Error("failed to generate", "question", q.TitleSlug, "err", err)
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

	if len(codeLinesToKeep) == 0 {
		return "", fmt.Errorf("no code found in %s", codeFile.Path)
	}

	return strings.Join(codeLinesToKeep, "\n"), nil
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
