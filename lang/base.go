package lang

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/charmbracelet/log"
	"github.com/dop251/goja"
	"github.com/spf13/viper"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

const (
	testCaseInputMark  = "input:"
	testCaseOutputMark = "output:"
	testCaseTargetMark = "target_case:"
)

type GenerateResult struct {
	Question *leetcode.QuestionData
	Lang     Lang
	OutDir   string
	SubDir   string
	Files    []FileOutput
	mask     int
}

type FileOutput struct {
	genResult *GenerateResult
	Filename  string
	Type      FileType
	Content   string
	Written   bool
}

func (f *FileOutput) GetPath() string {
	return filepath.Join(f.genResult.OutDir, f.genResult.SubDir, f.Filename)
}

func (f *FileOutput) GetContent() (string, error) {
	if f.Content == "" {
		content, err := os.ReadFile(f.GetPath())
		if err != nil {
			return "", err
		}
		f.Content = string(content)
	}
	return f.Content, nil
}

type FileType int

const (
	CodeFile FileType = 1 << iota
	TestFile
	TestCasesFile
	DocFile
	OtherFile
)

func (r *GenerateResult) AddFile(f FileOutput) *GenerateResult {
	if r.mask&int(f.Type) != 0 {
		panic(fmt.Sprintf("file type %d already exists", f.Type))
	}
	f.genResult = r
	r.Files = append(r.Files, f)
	r.mask |= int(f.Type)
	return r
}

func (r *GenerateResult) GetFile(typ FileType) *FileOutput {
	for _, f := range r.Files {
		if int(f.Type&typ) != 0 {
			return &f
		}
	}
	return nil
}

func (r *GenerateResult) SetOutDir(dir string) {
	r.OutDir = dir
}

type Lang interface {
	Name() string
	ShortName() string
	Slug() string
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

const codeContentTemplate = `
{{- block "header" . -}}
{{ .LineComment }} Created by {{ .Author }} at {{ .Time }}
{{ .LineComment }} {{ .Question.Url }}
{{ if .Question.IsContest }}{{ .LineComment }} {{ .Question.ContestUrl }}
{{ end }}
{{ end }}
{{ if not .SeparateDescriptionFile }}
{{ block "description" . -}}
{{ .BlockCommentStart }}
{{ block "title" . }}{{ .Question.QuestionFrontendId }}. {{ .Question.GetTitle }} ({{ .Question.Difficulty }}){{ end }}
{{ .Question.GetFormattedContent }}
{{ .BlockCommentEnd }}
{{ end }}
{{ end }}
{{ block "_internalBeforeMarker" . }}{{ end }}
{{ block "beforeMarker" . }}{{ end }}
{{ .LineComment }} {{ .CodeBeginMarker }}
{{ block "beforeCode" . }}{{ end }}
{{ block "code" . }}{{ .Code | runModifiers }}{{ end }}
{{ block "afterCode" . }}{{ end }}
{{ .LineComment }} {{ .CodeEndMarker }}
{{ block "afterMarker" . }}{{ end }}
{{ block "_internalAfterMarker" . }}{{ end }}
`

type codeContentData struct {
	Question                *leetcode.QuestionData
	Author                  string
	Time                    string
	LineComment             string
	BlockCommentStart       string
	BlockCommentEnd         string
	CodeBeginMarker         string
	CodeEndMarker           string
	Code                    string
	SeparateDescriptionFile bool
	NeedsDefinition         bool
}

var validBlocks = map[string]bool{
	"header":       true,
	"description":  true,
	"title":        true,
	"beforeMarker": true,
	"beforeCode":   true,
	"code":         true,
	"afterCode":    true,
	"afterMarker":  true,
}

// internal blocks are used to generate code for internal use.
const (
	internalBeforeMarker = "_internalBeforeMarker"
	internalAfterMarker  = "_internalAfterMarker"
)

var internalBlocks = map[string]bool{
	internalBeforeMarker: true,
	internalAfterMarker:  true,
}

var builtinModifiers = map[string]ModifierFunc{
	"removeUselessComments": removeUselessComments,
}

type ModifierFunc = func(string, *leetcode.QuestionData) string

func getBlocks(lang Lang) (ans []config.Block) {
	blocks := viper.Get("code." + lang.Slug() + ".blocks")
	if blocks == nil || len(blocks.([]any)) == 0 {
		blocks = viper.Get("code." + lang.ShortName() + ".blocks")
	}
	if blocks == nil || len(blocks.([]any)) == 0 {
		blocks = viper.Get("code.blocks")
	}
	if blocks == nil {
		return
	}
	for _, b := range blocks.([]any) {
		ans = append(
			ans, config.Block{
				Name:     b.(map[string]any)["name"].(string),
				Template: b.(map[string]any)["template"].(string),
			},
		)
	}
	return
}

func getModifiers(lang Lang, modifiersMap map[string]ModifierFunc) ([]ModifierFunc, error) {
	modifiers := viper.Get("code." + lang.Slug() + ".modifiers")
	if modifiers == nil || len(modifiers.([]any)) == 0 {
		modifiers = viper.Get("code." + lang.ShortName() + ".modifiers")
	}
	if modifiers == nil || len(modifiers.([]any)) == 0 {
		modifiers = viper.Get("code.modifiers")
	}
	if modifiers == nil {
		return nil, nil
	}

	var funcs []ModifierFunc
	for _, m := range modifiers.([]any) {
		m := m.(map[string]any)
		name, script := "", ""

		if m["name"] != nil {
			name = m["name"].(string)
			if f, ok := modifiersMap[name]; ok {
				funcs = append(funcs, f)
				continue
			}
		}
		if m["script"] != nil {
			script = m["script"].(string)
			vm := goja.New()
			_, err := vm.RunString(script)
			if err != nil {
				return nil, fmt.Errorf("failed to run script: %w", err)
			}
			var jsFn func(string) string
			if vm.Get("modify") == nil {
				return nil, fmt.Errorf("failed to get modify function")
			}
			err = vm.ExportTo(vm.Get("modify"), &jsFn)
			if err != nil {
				return nil, fmt.Errorf("failed to export function: %w", err)
			}
			f := func(s string, data *leetcode.QuestionData) string {
				return jsFn(s)
			}
			funcs = append(funcs, f)
			continue
		}
		log.Warn("invalid modifier, ignored", "name", name, "script", script)
	}
	return funcs, nil
}

func needsDefinition(code string) bool {
	return strings.Contains(code, "Definition for")
}

func needsMod(content string) bool {
	return strings.Contains(content, "<code>10<sup>9</sup> + 7</code>") || strings.Contains(content, "10^9 + 7")
}

func removeUselessComments(code string, q *leetcode.QuestionData) string {
	lines := strings.Split(code, "\n")
	var newLines []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "/**") && (strings.Contains(
			lines[i+1],
			"object will be instantiated and called",
		) || strings.Contains(lines[i+1], "Definition for")) {
			for {
				i++
				if strings.HasSuffix(lines[i], "*/") {
					break
				}
			}
			continue
		}
		newLines = append(newLines, line)
	}
	return strings.Join(newLines, "\n")
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

func (l baseLang) generateCodeContent(
	q *leetcode.QuestionData,
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
	_, err := tmpl.Parse(codeContentTemplate)
	if err != nil {
		return "", err
	}
	for _, block := range blocks {
		if !validBlocks[block.Name] && !internalBlocks[block.Name] {
			return "", fmt.Errorf("invalid block name: %s", block.Name)
		}
		_, err := tmpl.New(block.Name).Parse(block.Template)
		if err != nil {
			return "", err
		}
	}

	cfg := config.Get()
	data := &codeContentData{
		Question:                q,
		Author:                  cfg.Author,
		Time:                    time.Now().Format("2006/01/02 15:04"),
		LineComment:             l.lineComment,
		BlockCommentStart:       l.blockCommentStart,
		BlockCommentEnd:         l.blockCommentEnd,
		CodeBeginMarker:         config.CodeBeginMarker,
		CodeEndMarker:           config.CodeEndMarker,
		Code:                    code,
		SeparateDescriptionFile: separateDescriptionFile,
		NeedsDefinition:         needsDefinition(code),
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
	filename string,
	blocks []config.Block,
	modifiers []ModifierFunc,
	separateDescriptionFile bool,
) (
	FileOutput,
	error,
) {
	content, err := l.generateCodeContent(
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
		Type:     CodeFile,
	}, nil
}

func (l baseLang) generateTestCasesContent(q *leetcode.QuestionData) string {
	cases := q.GetTestCases()
	outputs := q.ParseExampleOutputs()
	argsNum := q.MetaData.NArg()

	// Assume all questions output are single.
	var caseAndOutputs []string
	for i := 0; i < len(cases) && i/argsNum < len(outputs); i += argsNum {
		input := strings.Join(cases[i:i+argsNum], "\n")
		caseAndOutputs = append(
			caseAndOutputs,
			fmt.Sprintf("%s\n%s\n%s\n%s", testCaseInputMark, input, testCaseOutputMark, outputs[i/argsNum]),
		)
	}
	content := strings.Join(caseAndOutputs, "\n\n")
	content = utils.EnsureTrailingNewline(content)
	return content
}

func (l baseLang) generateTestCasesFile(q *leetcode.QuestionData, filename string) (FileOutput, error) {
	content := l.generateTestCasesContent(q)
	content = fmt.Sprintf("%s 0\n\n", testCaseTargetMark) + content
	return FileOutput{
		Filename: filename,
		Content:  content,
		Type:     TestCasesFile,
	}, nil
}

// nolint: unused
func (l baseLang) generateTestFile(q *leetcode.QuestionData, filename string) (FileOutput, error) {
	return FileOutput{}, errors.New("not implemented")
}

func (l baseLang) generateDescriptionFile(q *leetcode.QuestionData, filename string) (FileOutput, error) {
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
		Filename: filename,
		Content:  content,
		Type:     DocFile,
	}, nil
}

func (l baseLang) GeneratePaths(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, l)
	baseFilename, err := q.GetFormattedFilename(l.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}

	genResult := &GenerateResult{
		Question: q,
		Lang:     l,
	}
	genResult.AddFile(
		FileOutput{
			Filename: baseFilename + l.extension,
			Type:     CodeFile,
		},
	)
	if separateDescriptionFile(l) {
		genResult.AddFile(
			FileOutput{
				Filename: baseFilename + ".md",
				Type:     DocFile,
			},
		)
	}
	return genResult, nil
}

func (l baseLang) Generate(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, l)
	baseFilename, err := q.GetFormattedFilename(l.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}

	genResult := &GenerateResult{
		Question: q,
		Lang:     l,
	}

	separateDescriptionFile := separateDescriptionFile(l)
	blocks := getBlocks(l)
	modifiers, err := getModifiers(l, builtinModifiers)
	if err != nil {
		return nil, err
	}
	codeFile, err := l.generateCodeFile(q, baseFilename+l.extension, blocks, modifiers, separateDescriptionFile)
	if err != nil {
		return nil, err
	}
	genResult.AddFile(codeFile)

	if separateDescriptionFile {
		docFile, err := l.generateDescriptionFile(q, baseFilename+".md")
		if err != nil {
			return nil, err
		}
		genResult.AddFile(docFile)
	}
	return genResult, nil
}
