package lang

type PythonConfig struct {
	Enable bool `json:"enable" yaml:"enable"`
}

type python struct {
	baseLang
}

func (p python) Name() string {
	return p.baseLang.Name
}

func (p python) Generate() []any {
	// TODO implement me
	panic("implement me")
}

func (p python) GenerateContest() []any {
	// TODO implement me
	panic("implement me")
}
