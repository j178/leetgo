package leetcode

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"text/template"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/PuerkitoBio/goquery"
	"github.com/goccy/go-json"
	"github.com/k3a/html2text"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/utils"
)

type TopicTag struct {
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	TranslatedName string `json:"translatedName"`
}

type CodeSnippet struct {
	LangSlug string `json:"langSlug"`
	Lang     string `json:"lang"`
	Code     string `json:"code"`
}

type Stats struct {
	TotalAccepted      string `json:"totalAccepted"`
	TotalSubmission    string `json:"totalSubmission"`
	TotalAcceptedRaw   int    `json:"totalAcceptedRaw"`
	TotalSubmissionRaw int    `json:"totalSubmissionRaw"`
	ACRate             string `json:"acRate"`
}

type statsNoMethods Stats

func (s *Stats) UnmarshalJSON(data []byte) error {
	// Cannot use `var v Stats` here, because it will cause infinite recursion.
	unquoted, err := strconv.Unquote(utils.BytesToString(data))
	if err != nil {
		unquoted = utils.BytesToString(data)
	}
	err = json.Unmarshal(utils.StringToBytes(unquoted), (*statsNoMethods)(s))
	if err != nil {
		return err
	}
	return nil
}

type MetaDataParam struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type MetaDataReturn struct {
	Type string `json:"type"`
	// Size    *int   `json:"size"`
	// ColSize *int `json:"colsize"`
	Dealloc bool `json:"dealloc"`
}

type MetaDataOutput struct {
	ParamIndex int `json:"paramindex"`
}

type MetaDataMethod struct {
	Name   string          `json:"name"`
	Params []MetaDataParam `json:"params"`
	Return MetaDataReturn  `json:"return"`
}

type MetaDataConstructor struct {
	Params []MetaDataParam `json:"params"`
}

// Normal problems metadata:
// {
//  "name": "minMovesToSeat",
//  "params": [
//    {
//      "name": "seats",
//      "type": "integer[]"
//    },
//    {
//      "type": "integer[]",
//      "name": "students"
//    }
//  ],
//  "return": {
//    "type": "integer"
//  }
// }

// System design problems metadata:
// {
//  "classname": "ExamRoom",
//  "maxbytesperline": 200000,
//  "constructor": {
//    "params": [
//      {
//        "type": "integer",
//        "name": "n"
//      }
//    ]
//  },
//  "methods": [
//    {
//      "name": "seat",
//      "params": [],
//      "return": {
//        "type": "integer"
//      }
//    },
//    {
//      "name": "leave",
//      "params": [
//        {
//          "type": "integer",
//          "name": "p"
//        }
//      ],
//      "return": {
//        "type": "void"
//      }
//    }
//  ],
//  "systemdesign": true,
//  "params": [
//    {
//      "name": "inputs",
//      "type": "integer[]"
//    },
//    {
//      "name": "inputs",
//      "type": "integer[]"
//    }
//  ],
//  "return": {
//    "type": "list<String>",
//    "dealloc": true
//  }
// }

type MetaData struct {
	Name   string          `json:"name"`
	Params []MetaDataParam `json:"params"`
	Return *MetaDataReturn `json:"return"`
	Output *MetaDataOutput `json:"output"`
	// System design problems related
	SystemDesign bool                `json:"systemdesign"`
	ClassName    string              `json:"classname"`
	Constructor  MetaDataConstructor `json:"constructor"`
	Methods      []MetaDataMethod    `json:"methods"`
	// Unknown fields
	Manual bool `json:"manual"`
}

type metaDataNoMethods MetaData

// Type name in metadata is not consistent, we need to normalize it.
func normalizeType(ty string) string {
	switch {
	case strings.HasPrefix(ty, "list<"):
		return normalizeType(ty[5:len(ty)-1]) + "[]" // "list<int>" -> "int[]"
	case ty == "String":
		return "string"
	case ty == "":
		return "void"
	}
	return ty
}

func (m *MetaData) normalize() {
	for i, param := range m.Params {
		m.Params[i].Type = normalizeType(param.Type)
	}
	if m.Return != nil {
		m.Return.Type = normalizeType(m.Return.Type)
	}
	for _, method := range m.Methods {
		for i, param := range method.Params {
			method.Params[i].Type = normalizeType(param.Type)
		}
		method.Return.Type = normalizeType(method.Return.Type)
	}
}

func (m *MetaData) UnmarshalJSON(data []byte) error {
	// Ignore error, when we load from sqlite, no need to unquote it.
	unquoted, err := strconv.Unquote(utils.BytesToString(data))
	if err != nil {
		unquoted = utils.BytesToString(data)
	}
	err = json.Unmarshal(utils.StringToBytes(unquoted), (*metaDataNoMethods)(m))
	if err != nil {
		return err
	}
	m.normalize()
	return nil
}

func (m *MetaData) NArg() int {
	if m.SystemDesign {
		return 2
	}
	return len(m.Params)
}

func (m *MetaData) ResultType() string {
	if m.Return != nil && m.Return.Type != "void" {
		return m.Return.Type
	} else {
		return m.Params[m.Output.ParamIndex].Type
	}
}

type JsonExampleTestCases []string

type jsonExampleTestCasesNoMethods JsonExampleTestCases

func (j *JsonExampleTestCases) UnmarshalJSON(data []byte) error {
	unquoted, err := strconv.Unquote(utils.BytesToString(data))
	if err != nil {
		unquoted = utils.BytesToString(data)
	}
	err = json.Unmarshal(utils.StringToBytes(unquoted), (*jsonExampleTestCasesNoMethods)(j))
	if err != nil {
		return err
	}
	return nil
}

type SimilarQuestion struct {
	Title           string `json:"title"`
	TitleSlug       string `json:"titleSlug"`
	Difficulty      string `json:"difficulty"`
	TranslatedTitle string `json:"translatedTitle"`
}

type SimilarQuestions []SimilarQuestion

type similarQuestionsNoMethods SimilarQuestions

func (s *SimilarQuestions) UnmarshalJSON(data []byte) error {
	unquoted, err := strconv.Unquote(utils.BytesToString(data))
	if err != nil {
		unquoted = utils.BytesToString(data)
	}
	err = json.Unmarshal(utils.StringToBytes(unquoted), (*similarQuestionsNoMethods)(s))
	if err != nil {
		return err
	}
	return nil
}

type CategoryTitle string

const (
	CategoryAlgorithms  CategoryTitle = "Algorithms"
	CategoryDatabase    CategoryTitle = "Database"
	CategoryShell       CategoryTitle = "Shell"
	CategoryConcurrency CategoryTitle = "Concurrency"
	CategoryAll         CategoryTitle = ""
)

type QuestionData struct {
	client               Client
	contest              *Contest
	partial              int32
	TitleSlug            string               `json:"titleSlug"`
	QuestionId           string               `json:"questionId"`
	QuestionFrontendId   string               `json:"questionFrontendId"`
	CategoryTitle        CategoryTitle        `json:"categoryTitle"`
	Title                string               `json:"title"`
	TranslatedTitle      string               `json:"translatedTitle"`
	Difficulty           string               `json:"difficulty"`
	TopicTags            []TopicTag           `json:"topicTags"`
	IsPaidOnly           bool                 `json:"isPaidOnly"`
	Content              string               `json:"content"`
	TranslatedContent    string               `json:"translatedContent"`
	Status               string               `json:"status"` // "ac", "notac", or null
	Stats                Stats                `json:"stats"`
	Hints                []string             `json:"hints"`
	SimilarQuestions     SimilarQuestions     `json:"similarQuestions"`
	SampleTestCase       string               `json:"sampleTestCase"`
	ExampleTestcases     string               `json:"exampleTestcases"`
	JsonExampleTestcases JsonExampleTestCases `json:"jsonExampleTestcases"`
	ExampleTestcaseList  []string             `json:"exampleTestcaseList"`
	MetaData             MetaData             `json:"metaData"`
	CodeSnippets         []CodeSnippet        `json:"codeSnippets"`
}

func (q *QuestionData) Url() string {
	return q.client.BaseURI() + "problems/" + q.TitleSlug + "/"
}

func (q *QuestionData) ContestUrl() string {
	return q.client.BaseURI() + "contest/" + q.contest.TitleSlug + "/problems/" + q.TitleSlug + "/"
}

func (q *QuestionData) IsContest() bool {
	return q.contest != nil
}

func (q *QuestionData) Contest() *Contest {
	return q.contest
}

func (q *QuestionData) Fulfill() (err error) {
	if atomic.LoadInt32(&q.partial) == 0 {
		return
	}

	contest := q.contest
	var nq *QuestionData
	if q.IsContest() {
		nq, err = q.client.GetContestQuestionData(q.contest.TitleSlug, q.TitleSlug)
	} else {
		nq, err = q.client.GetQuestionData(q.TitleSlug)
	}
	if err != nil {
		return
	}
	*q = *nq
	q.contest = contest
	atomic.StoreInt32(&q.partial, 0)
	return nil
}

func (q *QuestionData) GetTitle() string {
	if config.Get().Language == config.ZH && q.TranslatedTitle != "" {
		return q.TranslatedTitle
	}
	return q.Title
}

func (q *QuestionData) GetContent() (string, config.Language) {
	if config.Get().Language == config.ZH && q.TranslatedContent != "" {
		return q.TranslatedContent, config.ZH
	}
	if config.Get().Language == config.EN && (q.Content == "" || strings.Contains(
		q.Content,
		"English description is not available for the problem.",
	)) {
		return q.TranslatedContent, config.ZH
	}
	return q.Content, config.EN
}

func (q *QuestionData) GetFormattedContent() string {
	content, lang := q.GetContent()

	// Convert to markdown
	converter := md.NewConverter("", true, nil)
	converter.Use(plugin.GitHubFlavored())
	replaceSub := md.Rule{
		Filter: []string{"sub"},
		Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
			selec.SetText(utils.ReplaceSubscript(content))
			return nil
		},
	}
	replaceSup := md.Rule{
		Filter: []string{"sup"},
		Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
			selec.SetText(utils.ReplaceSuperscript(content))
			return nil
		},
	}
	replaceEm := md.Rule{
		Filter: []string{"em"},
		Replacement: func(content string, selec *goquery.Selection, options *md.Options) *string {
			return md.String(content)
		},
	}
	converter.AddRules(replaceSub, replaceSup, replaceEm)
	content, err := converter.ConvertString(content)
	if err != nil {
		return content
	}

	// Remove special HTML entities characters
	replacer := strings.NewReplacer("\u00A0", " ", "\u200B", "")
	content = replacer.Replace(content)

	// Wrap and remove blank lines
	if lang == config.EN {
		content = wordwrap.String(content, 100)
	} else {
		content = wrap.String(content, 60)
	}
	content = utils.CondenseEmptyLines(content)
	content = utils.EnsureTrailingNewline(content)
	return content
}

var (
	enPat = regexp.MustCompile(`<strong>Output[:：]?\s?</strong>\s?\n?\s*(.+)`)
	zhPat = regexp.MustCompile(`<strong>输出[:：]?\s?</strong>\s?\n?\s*(.+)`)
)

func (q *QuestionData) GetTestCases() []string {
	var cases []string
	if len(q.JsonExampleTestcases) > 0 {
		for _, c := range q.JsonExampleTestcases {
			cases = append(cases, strings.Split(c, "\n")...)
		}
	} else if len(q.ExampleTestcaseList) > 0 {
		for _, c := range q.ExampleTestcaseList {
			cases = append(cases, strings.Split(c, "\n")...)
		}
	} else if q.ExampleTestcases != "" {
		cases = strings.Split(q.ExampleTestcases, "\n")
	} else if q.SampleTestCase != "" {
		cases = strings.Split(q.SampleTestCase, "\n")
	}
	return cases
}

// ParseExampleOutputs parses example output from content and translatedContent.
// We can also get correct example outputs by submitting example inputs to judge server.
func (q *QuestionData) ParseExampleOutputs() []string {
	var pat *regexp.Regexp
	var content string
	if q.Content != "" && !strings.Contains(q.Content, "English description is not available for the problem.") {
		content = q.Content
		pat = enPat
	} else {
		content = q.TranslatedContent
		pat = zhPat
	}
	found := pat.FindAllStringSubmatch(content, -1)
	result := make([]string, 0, len(found))
	// TODO multi-line output, like https://leetcode.cn/problems/find-valid-matrix-given-row-and-column-sums/
	for _, f := range found {
		output := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(f[1]), "<code>"), "</pre>")
		output = html2text.HTMLEntitiesToText(output)
		output = strings.ReplaceAll(output, ", ", ",")
		result = append(result, output)
	}
	return result
}

func (q *QuestionData) TagSlugs() []string {
	slugs := make([]string, 0, len(q.TopicTags))
	for _, tag := range q.TopicTags {
		slugs = append(slugs, tag.Slug)
	}
	return slugs
}

func (q *QuestionData) GetCodeSnippet(slug string) string {
	for _, snippet := range q.CodeSnippets {
		if slug == snippet.LangSlug {
			return snippet.Code
		}
	}
	return ""
}

type FilenameTemplateData struct {
	Id               string
	Slug             string
	Title            string
	Difficulty       string
	Lang             string
	SlugIsMeaningful bool
	IsContest        bool
	ContestTitle     string
	ContestShortSlug string
	ContestSlug      string
}

func (q *QuestionData) formatQuestionId() (string, bool) {
	slugValid := true
	id := q.QuestionFrontendId
	switch {
	case strings.HasPrefix(id, "剑指 Offer"):
		slugValid = false
		cid := strings.TrimSpace(id[len("剑指 Offer")+1:])
		cid = strings.ReplaceAll(cid, " ", "-")
		cid = strings.ReplaceAll(cid, "---", "-")
		id = "Offer-" + cid
	case strings.HasPrefix(id, "面试题"):
		slugValid = false
		cid := strings.TrimSpace(id[len("面试题")+1:])
		cid = strings.ReplaceAll(cid, " ", "-")
		id = "Interview-" + cid
	case strings.HasPrefix(id, "LCP"), strings.HasPrefix(id, "LCS"):
		slugValid = false
		id = strings.ReplaceAll(id, " ", "-")
	}
	return id, slugValid
}

func contestShortSlug(contestSlug string) string {
	return strings.Replace(contestSlug, "-contest-", "-", 1)
}

func (q *QuestionData) GetFormattedFilename(lang string, filenameTemplate string) (string, error) {
	id, slugValid := q.formatQuestionId()
	data := &FilenameTemplateData{
		Id:               id,
		Slug:             q.TitleSlug,
		Title:            q.GetTitle(),
		Difficulty:       q.Difficulty,
		Lang:             lang,
		SlugIsMeaningful: slugValid,
		IsContest:        q.IsContest(),
	}
	if q.IsContest() {
		// Override id with contest question number
		id, err := q.contest.GetQuestionNumber(q.TitleSlug)
		if err != nil {
			panic(fmt.Errorf("failed to get question number for %s: %w", q.TitleSlug, err))
		}
		data.Id = strconv.Itoa(id)
		data.ContestSlug = q.contest.TitleSlug
		data.ContestTitle = q.contest.Title
		data.ContestShortSlug = contestShortSlug(q.contest.TitleSlug)
	}
	tmpl := template.New("filename")
	tmpl.Funcs(
		template.FuncMap{
			"lower": strings.ToLower,
			"upper": strings.ToUpper,
			"trim":  strings.TrimSpace,
			"padWithZero": func(n int, s string) string {
				return fmt.Sprintf("%0*s", n, s)
			},
			"toUnderscore": func(s string) string {
				return strings.ReplaceAll(s, "-", "_")
			},
		},
	)
	tmpl, err := tmpl.Parse(filenameTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
