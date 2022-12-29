package lang

type GoConfig struct {
	Enable           bool   `json:"enable" yaml:"enable"`
	SeparatePackage  bool   `json:"separate_package" yaml:"separate_package"`
	FilenameTemplate string `json:"filename_template" yaml:"filename_template"`
}

type golang struct {
	baseLang
}

func (golang) Name() string {
	return "Go"
}

func (golang) Generate() []any {
	return nil
}

func (golang) GenerateContest() []any {
	return nil
}
