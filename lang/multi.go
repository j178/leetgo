package lang

import (
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
)

type MultiGenerator struct {
	generators []Generator
}

func NewMultiGenerator() MultiGenerator {
	cfg := config.Get()
	var gen []Generator
	if cfg.Go.Enable {
		gen = append(gen, golangGen)
	}
	if cfg.Python.Enable {
		gen = append(gen, pythonGen)
	}
	return MultiGenerator{generators: gen}
}

func (m MultiGenerator) Generate(q leetcode.QuestionData) ([][]FileOutput, error) {
	var files [][]FileOutput
	for _, gen := range m.generators {
		files = append(files, gen.Generate(q))
	}
	return files, nil
}
