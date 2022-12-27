package lang

type Lang interface {
	Name() string
	Ext() string
}

var SupportedLanguages = []Lang{
	golang{},
}
