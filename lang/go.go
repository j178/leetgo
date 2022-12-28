package lang

type GoConfig struct {
    SeparatePackage  bool   `json:"separate_package"`
    FilenameTemplate string `json:"filename_template"`
}

type golang struct {
}

func (golang) Name() string {
    return "Go"
}

func (golang) Ext() string {
    return "go"
}
