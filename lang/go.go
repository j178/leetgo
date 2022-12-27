package lang

type GoConfig struct {
	SeparatePackage  bool
	FilenameTemplate string
}

type golang struct {
}

func (golang) Name() string {
	return "Go"
}

func (golang) Ext() string {
	return "go"
}
