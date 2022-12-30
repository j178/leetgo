package lang

import (
	"errors"
	"fmt"

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

var NotSupported = errors.New("not supported")

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
	supportedLanguages = []Generator{
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
	shortNames = func() map[string]Generator {
		m := make(map[string]Generator)
		for _, g := range supportedLanguages {
			m[g.ShortName()] = g
		}
		return m
	}()
)

func Generate(q leetcode.QuestionData) ([][]FileOutput, error) {
	cfg := config.Get()
	var files [][]FileOutput
	gen, ok := shortNames[cfg.Gen]
	if !ok {
		return nil, fmt.Errorf("language %s is not supported yet", cfg.Gen)
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
