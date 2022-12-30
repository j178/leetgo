package lang

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
)

type baseLang struct {
	Name              string
	ShortName         string
	Extension         string
	LineComment       string
	BlockCommentStart string
	BlockCommentEnd   string
}

type FileOutput struct {
	Filename string
	Content  string
}

var NotSupported = errors.New("not supported")

type Generator interface {
	Name() string
	ShortName() string
	Generate(q leetcode.QuestionData) ([]FileOutput, error)
	GenerateTest(q leetcode.QuestionData) ([]FileOutput, error)
	GenerateContest(c leetcode.Contest) ([]FileOutput, error)
	GenerateContestTest(c leetcode.Contest) ([]FileOutput, error)
}

var (
	cppGen = commonGenerator{
		baseLang: baseLang{
			Name:              "C++",
			ShortName:         "cpp",
			Extension:         ".cpp",
			LineComment:       "//",
			BlockCommentStart: "/*",
			BlockCommentEnd:   "*/",
		},
	}
	rustGen = commonGenerator{
		baseLang: baseLang{
			Name:              "Rust",
			ShortName:         "rs",
			Extension:         ".rs",
			LineComment:       "//",
			BlockCommentStart: "/*",
			BlockCommentEnd:   "*/",
		},
	}
	javaGen = commonGenerator{
		baseLang: baseLang{
			Name:              "Java",
			ShortName:         "java",
			Extension:         ".java",
			LineComment:       "//",
			BlockCommentStart: "/*",
			BlockCommentEnd:   "*/",
		},
	}

	supportedLanguages = map[string]Generator{
		golangGen.ShortName(): golangGen,
		pythonGen.ShortName(): pythonGen,
		cppGen.ShortName():    cppGen,
		rustGen.ShortName():   rustGen,
		javaGen.ShortName():   javaGen,
	}
)

func Generate(q leetcode.QuestionData) ([][]FileOutput, error) {
	cfg := config.Get()
	var files [][]FileOutput
	gen, ok := supportedLanguages[cfg.Gen]
	if !ok {
		return nil, fmt.Errorf("language %s is not supported yet", cfg.Gen)
	}
	f, err := gen.Generate(q)
	if err != nil {
		return nil, err
	}
	files = append(files, f)
	f, err = gen.GenerateTest(q)
	if err != nil {
		if err == NotSupported {
			hclog.L().Warn("test generation not supported for language, skip", "language", gen.Name())
		}
	}
	files = append(files, f)

	return files, nil
}

type commonGenerator struct {
	baseLang
}

func (g commonGenerator) Name() string {
	return g.baseLang.Name
}

func (g commonGenerator) ShortName() string {
	return g.baseLang.ShortName
}

func (g commonGenerator) Generate(q leetcode.QuestionData) ([]FileOutput, error) {
	return nil, NotSupported
}

func (g commonGenerator) GenerateTest(q leetcode.QuestionData) ([]FileOutput, error) {
	return nil, NotSupported
}

func (g commonGenerator) GenerateContest(c leetcode.Contest) ([]FileOutput, error) {
	return nil, NotSupported
}

func (g commonGenerator) GenerateContestTest(c leetcode.Contest) ([]FileOutput, error) {
	return nil, NotSupported
}
