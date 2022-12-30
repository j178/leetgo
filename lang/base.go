package lang

import (
	"errors"

	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
)

type baseLang struct {
	Name              string
	ShortName         string
	Suffix            string
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

var SupportedLanguages = []Generator{
	golangGen,
	pythonGen,
}

func Generate(q leetcode.QuestionData) ([][]FileOutput, error) {
	cfg := config.Get()
	var files [][]FileOutput
	var gen Generator
	switch {
	case cfg.Go.Enable:
		gen = golangGen
	case cfg.Python.Enable:
		gen = pythonGen
	}
	if gen == nil {
		return nil, errors.New("no language enabled")
	}
	f, err := gen.Generate(q)
	if err != nil {
		return nil, err
	}
	files = append(files, f)
	f, err = gen.GenerateTest(q)
	if err != nil {
		if err == NotSupported {
			hclog.L().Warn("test generation not supported for language, skip", "lang", gen.Name())
		}
	}

	return files, nil
}
