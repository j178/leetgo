package lang

type baseLang struct {
    Name              string
    Suffix            string
    LineComment       string
    BlockCommentStart string
    BlockCommentEnd   string
}

type LangGenerator interface {
    Name() string
    Generate() []any
    GenerateContest() []any
}

var SupportedLanguages = []LangGenerator{
    golang{
        baseLang{
            Name:              "go",
            Suffix:            ".go",
            LineComment:       "//",
            BlockCommentStart: "/*",
            BlockCommentEnd:   "*/",
        },
    },
}
