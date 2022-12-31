package lang

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/viper"
)

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

// TODO use template
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

func (l baseLang) generateCode(q *leetcode.QuestionData, modifiers ...func(string) string) string {
	code := q.GetCodeSnippet(l.Slug())
	for _, m := range modifiers {
		code = m(code)
	}
	return code
}

func addCodeMark(comment string) func(string) string {
	return func(s string) string {
		return fmt.Sprintf("%s %s\n\n%s\n\n%s %s", comment, config.CodeBeginMark, s, comment, config.CodeEndMark)
	}
}

func (l baseLang) Generate(q *leetcode.QuestionData) ([]FileOutput, error) {
	comment := l.generateComments(q)
	code := l.generateCode(q, addCodeMark(l.lineComment))
	content := comment + "\n" + code + "\n"

	files := FileOutput{
		// TODO filename template
		Filename: fmt.Sprintf("%s%s", q.TitleSlug, l.extension),
		Content:  content,
	}
	return []FileOutput{files}, nil
}

func (l baseLang) GenerateTest(q *leetcode.QuestionData) ([]FileOutput, error) {
	// 检查基础库是否生成，如果没有生成，先生成基础库
	return nil, NotSupported
}

type FileOutput struct {
	BaseDir  string
	Filename string
	Content  string
}

var (
	NotSupported   = errors.New("not supported")
	NotImplemented = errors.New("not implemented")
)

type Generator interface {
	Name() string
	ShortName() string
	Slug() string
	Generate(q *leetcode.QuestionData) ([]FileOutput, error)
	GenerateTest(q *leetcode.QuestionData) ([]FileOutput, error)
}

func getGenerator(gen string) Generator {
	gen = strings.ToLower(gen)
	for _, l := range SupportedLanguages {
		if strings.HasPrefix(l.ShortName(), gen) || strings.HasPrefix(l.Slug(), gen) {
			return l
		}
	}
	return nil
}

func Generate(q *leetcode.QuestionData) ([]FileOutput, error) {
	cfg := config.Get()
	var files []FileOutput
	gen := getGenerator(cfg.Gen)
	if gen == nil {
		return nil, fmt.Errorf("language %s is not supported yet, welcome to send a PR", cfg.Gen)
	}

	codeSnippet := q.GetCodeSnippet(gen.Slug())
	if codeSnippet == "" {
		return nil, fmt.Errorf("no %s code snippet found for %s", cfg.Gen, q.TitleSlug)
	}

	f, err := gen.Generate(q)
	if err != nil {
		return nil, err
	}
	files = append(files, f...)
	f, err = gen.GenerateTest(q)
	if err != nil {
		if err == NotSupported {
			hclog.L().Warn("test generation not supported for language, skip", "language", gen.Name())
		}
	} else {
		files = append(files, f...)
	}

	dir := viper.GetString(cfg.Gen + ".out_dir")
	if dir == "" {
		dir = cfg.Gen
	}
	for i := range files {
		files[i].BaseDir = dir
	}
	return files, nil
}
