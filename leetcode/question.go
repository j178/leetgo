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
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/utils"
	"github.com/k3a/html2text"
	"github.com/mitchellh/go-wordwrap"
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

func (s *Stats) UnmarshalJSON(data []byte) error {
	// Cannot use `var v Stats` here, because it will cause infinite recursion.
	unquoted, err := strconv.Unquote(utils.BytesToString(data))
	if err != nil {
		unquoted = utils.BytesToString(data)
	}
	type alias Stats
	var v alias
	if err := json.Unmarshal(utils.StringToBytes(unquoted), &v); err != nil {
		return err
	}
	*s = Stats(v)
	return nil
}

type MetaDataParam struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type MetaDataReturn struct {
	Type    string `json:"type"`
	Size    int    `json:"size"`
	Dealloc bool   `json:"dealloc"`
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
	Return MetaDataReturn  `json:"return"`
	// System design problems related
	SystemDesign bool                `json:"systemdesign"`
	ClassName    string              `json:"classname"`
	Constructor  MetaDataConstructor `json:"constructor"`
	Methods      []MetaDataMethod    `json:"methods"`
	// Unknown fields
	Manual bool `json:"manual"`
}

func (m *MetaData) UnmarshalJSON(data []byte) error {
	// Ignore error, when we loads from sqlite, no need to unquote it.
	unquoted, err := strconv.Unquote(utils.BytesToString(data))
	if err != nil {
		unquoted = utils.BytesToString(data)
	}
	type alias MetaData
	var v alias
	if err := json.Unmarshal(utils.StringToBytes(unquoted), &v); err != nil {
		return err
	}
	*m = MetaData(v)
	return nil
}

type JsonExampleTestCases []string

func (j *JsonExampleTestCases) UnmarshalJSON(data []byte) error {
	unquoted, err := strconv.Unquote(utils.BytesToString(data))
	if err != nil {
		unquoted = utils.BytesToString(data)
	}
	var v []any
	if err := json.Unmarshal(utils.StringToBytes(unquoted), &v); err != nil {
		return err
	}
	for _, c := range v {
		*j = append(*j, c.(string))
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

func (s *SimilarQuestions) UnmarshalJSON(data []byte) error {
	unquoted, err := strconv.Unquote(utils.BytesToString(data))
	if err != nil {
		unquoted = utils.BytesToString(data)
	}
	type alias SimilarQuestions
	var v alias
	if err := json.Unmarshal(utils.StringToBytes(unquoted), &v); err != nil {
		return err
	}
	*s = SimilarQuestions(v)
	return nil
}

type QuestionData struct {
	client               Client
	contest              *Contest
	partial              int32
	TitleSlug            string               `json:"titleSlug"`
	QuestionId           string               `json:"questionId"`
	QuestionFrontendId   string               `json:"questionFrontendId"`
	CategoryTitle        string               `json:"categoryTitle"`
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

func (q *QuestionData) Fulfill() error {
	if atomic.LoadInt32(&q.partial) == 0 {
		return nil
	}

	// TODO 为 contest 适配
	q1, err := q.client.GetQuestionData(q.TitleSlug)
	if err != nil {
		return err
	}
	*q = *q1
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
	maxWidth := uint(100)
	if lang == config.ZH {
		maxWidth = 60
	}
	content = wordwrap.WrapString(content, maxWidth)
	content = utils.RemoveEmptyLine(content)
	return content
}

var (
	enPat = regexp.MustCompile(`<strong>Output[:：]?\s?</strong>\s?\n?\s*(.+)`)
	zhPat = regexp.MustCompile(`<strong>输出[:：]?\s?</strong>\s?\n?\s*(.+)`)
)

func (q *QuestionData) GetTestCases() []string {
	var cases []string
	if len(q.JsonExampleTestcases) > 0 {
		cases = q.JsonExampleTestcases
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
	for _, f := range found {
		output := strings.TrimSuffix(strings.TrimSpace(f[1]), "</pre>")
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
		id = "剑指Offer-" + cid
	case strings.HasPrefix(id, "面试题"):
		slugValid = false
		id = strings.ReplaceAll(id, " ", "-")
	case strings.HasPrefix(id, "LCP"), strings.HasPrefix(id, "LCS"):
		slugValid = false
		id = strings.ReplaceAll(id, " ", "-")
	}
	return id, slugValid
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
	}
	tmpl := template.New("filename")
	tmpl.Funcs(
		template.FuncMap{
			"lower": strings.ToLower,
			"upper": strings.ToUpper,
			"trim":  strings.TrimSpace,
			"padWithZero": func(n int, s string) string {
				return fmt.Sprintf("%0"+strconv.Itoa(n)+"s", s)
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
