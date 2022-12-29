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
		gen = append(gen, golang{})
	}
	if cfg.Python.Enable {
		gen = append(gen, python{})
	}
	return MultiGenerator{generators: gen}
}

func (m MultiGenerator) Generate(q leetcode.QuestionData) error {
	for _, gen := range m.generators {
		gen.Generate(q)
	}
	return nil
}
