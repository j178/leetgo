package lang

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/viper"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

func getCodeStringConfig(lang Lang, key string) string {
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
	ans := getCodeStringConfig(gen, "filename_template")
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
	outDir := getCodeStringConfig(lang, "out_dir")
	// If outDir is not set, use the language slug as the outDir.
	if outDir == "" {
		outDir = lang.Slug()
	}
	outDir = filepath.Join(cfg.ProjectRoot(), outDir)
	return outDir
}

func separateDescriptionFile(lang Lang) bool {
	ans := viper.Get("code." + lang.Slug() + ".separate_description_file")
	if ans != nil {
		return ans.(bool)
	}
	ans = viper.Get("code." + lang.ShortName() + ".separate_description_file")
	if ans != nil {
		return ans.(bool)
	}
	return config.Get().Code.SeparateDescriptionFile
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

func (l baseLang) generateCodeContent(
	q *leetcode.QuestionData,
	baseFilename string,
	blocks []config.Block,
	modifiers []ModifierFunc,
	separateDescriptionFile bool,
) (string, error) {
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
	var err error
	if separateDescriptionFile {
		_, err = tmpl.Parse(withoutDescriptionContentTemplate)
	} else {
		_, err = tmpl.Parse(defaultContentTemplate)
	}
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
	content := buf.String()
	content = utils.CondenseEmptyLines(content)
	content = utils.EnsureTrailingNewline(content)
	return content, nil
}

func (l baseLang) generateCodeFile(
	q *leetcode.QuestionData,
	baseFilename string,
	blocks []config.Block,
	modifiers []ModifierFunc,
	separateDescriptionFile bool,
) (
	FileOutput,
	error,
) {
	content, err := l.generateCodeContent(
		q,
		baseFilename,
		blocks,
		modifiers,
		separateDescriptionFile,
	)
	if err != nil {
		return FileOutput{}, err
	}
	return FileOutput{
		Path:    baseFilename + l.extension,
		Content: content,
		Type:    CodeFile,
	}, nil
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

func (l baseLang) generateDescriptionFile(q *leetcode.QuestionData, baseFilename string) (FileOutput, error) {
	tmpl := `# [%s. %s](%s) (%s)
%s`
	url := ""
	if q.IsContest() {
		url = q.ContestUrl()
	} else {
		url = q.Url()
	}
	content := fmt.Sprintf(
		tmpl,
		q.QuestionFrontendId,
		q.GetTitle(),
		url,
		q.Difficulty,
		q.GetFormattedContent(),
	)
	return FileOutput{
		Path:    baseFilename + ".md",
		Content: content,
		Type:    DocFile,
	}, nil
}

func (l baseLang) GeneratePaths(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, l)
	baseFilename, err := q.GetFormattedFilename(l.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}

	code := FileOutput{
		Path: baseFilename + l.extension,
		Type: CodeFile,
	}
	files := []FileOutput{code}

	if separateDescriptionFile(l) {
		files = append(
			files, FileOutput{
				Path: baseFilename + ".md",
				Type: DocFile,
			},
		)
	}

	return &GenerateResult{
		Question: q,
		Lang:     l,
		Files:    files,
	}, nil
}

func (l baseLang) Generate(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, l)
	baseFilename, err := q.GetFormattedFilename(l.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}

	separateDescriptionFile := separateDescriptionFile(l)
	blocks := getBlocks(l)
	modifiers, err := getModifiers(l, builtinModifiers)
	if err != nil {
		return nil, err
	}
	codeFile, err := l.generateCodeFile(q, baseFilename, blocks, modifiers, separateDescriptionFile)
	if err != nil {
		return nil, err
	}
	files := []FileOutput{codeFile}

	if separateDescriptionFile {
		docFile, err := l.generateDescriptionFile(q, baseFilename)
		if err != nil {
			return nil, err
		}
		files = append(files, docFile)
	}
	return &GenerateResult{
		Question: q,
		Lang:     l,
		Files:    files,
	}, nil
}
