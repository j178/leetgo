package lang

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
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

type FileOutput struct {
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
	Generate(q leetcode.QuestionData) ([]FileOutput, error)
	GenerateTest(q leetcode.QuestionData) ([]FileOutput, error)
	GenerateContest(c leetcode.Contest) ([]FileOutput, error)
	GenerateContestTest(c leetcode.Contest) ([]FileOutput, error)
}

var (
	golangGen = golang{
		baseLang{
			name:              "Go",
			slug:              "golang",
			shortName:         "go",
			extension:         ".go",
			lineComment:       "//",
			blockCommentStart: "/*",
			blockCommentEnd:   "*/",
		},
	}
	pythonGen = python{
		baseLang{
			name:              "Python",
			slug:              "python",
			shortName:         "py",
			extension:         ".py",
			lineComment:       "#",
			blockCommentStart: `"""`,
			blockCommentEnd:   `"""`,
		},
	}
	cppGen = commonGenerator{
		baseLang: baseLang{
			name:              "C++",
			slug:              "cpp",
			shortName:         "cpp",
			extension:         ".cpp",
			lineComment:       "//",
			blockCommentStart: "/*",
			blockCommentEnd:   "*/",
		},
	}
	rustGen = commonGenerator{
		baseLang: baseLang{
			name:              "Rust",
			slug:              "rust",
			shortName:         "rs",
			extension:         ".rs",
			lineComment:       "//",
			blockCommentStart: "/*",
			blockCommentEnd:   "*/",
		},
	}
	javaGen = commonGenerator{
		baseLang: baseLang{
			name:              "Java",
			slug:              "java",
			shortName:         "java",
			extension:         ".java",
			lineComment:       "//",
			blockCommentStart: "/*",
			blockCommentEnd:   "*/",
		},
	}
	cGen = commonGenerator{
		baseLang: baseLang{
			name:              "C",
			slug:              "c",
			shortName:         "c",
			extension:         ".c",
			lineComment:       "//",
			blockCommentStart: "/*",
			blockCommentEnd:   "*/",
		},
	}
	csharpGen = commonGenerator{
		baseLang: baseLang{
			name:              "C#",
			slug:              "csharp",
			shortName:         "cs",
			extension:         ".cs",
			lineComment:       "//",
			blockCommentStart: "/*",
			blockCommentEnd:   "*/",
		},
	}
	jsGen = commonGenerator{
		baseLang: baseLang{
			name:              "JavaScript",
			slug:              "javascript",
			shortName:         "js",
			extension:         ".js",
			lineComment:       "//",
			blockCommentStart: "/*",
			blockCommentEnd:   "*/",
		},
	}
	rubyGen = commonGenerator{
		baseLang: baseLang{
			name:              "Ruby",
			slug:              "ruby",
			shortName:         "rb",
			extension:         ".rb",
			lineComment:       "#",
			blockCommentStart: "=begin",
			blockCommentEnd:   "=end",
		},
	}
	swiftGen = commonGenerator{
		baseLang: baseLang{
			name:              "Swift",
			slug:              "swift",
			shortName:         "swift",
			extension:         ".swift",
			lineComment:       "//",
			blockCommentStart: "/*",
			blockCommentEnd:   "*/",
		},
	}
	kotlinGen = commonGenerator{
		baseLang: baseLang{
			name:              "Kotlin",
			slug:              "kotlin",
			shortName:         "kt",
			extension:         ".kt",
			lineComment:       "//",
			blockCommentStart: "/*",
			blockCommentEnd:   "*/",
		},
	}
	// TODO scala, typescript, php, erlang, dart, racket
	SupportedLanguages = []Generator{
		golangGen,
		pythonGen,
		cppGen,
		rustGen,
		javaGen,
		cGen,
		csharpGen,
		jsGen,
		rubyGen,
		swiftGen,
		kotlinGen,
	}
)

func getGenerator(gen string) Generator {
	gen = strings.ToLower(gen)
	for _, l := range SupportedLanguages {
		if strings.HasPrefix(l.ShortName(), gen) || strings.HasPrefix(l.Slug(), gen) {
			return l
		}
	}
	return nil
}

func Generate(q leetcode.QuestionData) ([][]FileOutput, error) {
	cfg := config.Get()
	var files [][]FileOutput
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

func (g commonGenerator) Generate(q leetcode.QuestionData) ([]FileOutput, error) {
	return nil, NotImplemented
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
