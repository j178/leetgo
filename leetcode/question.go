package leetcode

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/utils"
	"github.com/jedib0t/go-pretty/v6/text"
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
		return err
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
	unquoted, err := strconv.Unquote(utils.BytesToString(data))
	if err != nil {
		return err
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
		return err
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
		return err
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
	contestSlug          string
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
	return q.client.BaseURI() + "contest/" + q.contestSlug + "/problems/" + q.TitleSlug + "/"
}

func (q *QuestionData) IsContest() bool {
	return q.contestSlug != ""
}

func (q *QuestionData) GetTitle() string {
	if config.Get().Language == config.ZH && q.TranslatedTitle != "" {
		return q.TranslatedTitle
	}
	return q.Title
}

func (q *QuestionData) GetContent() string {
	if config.Get().Language == config.ZH && q.TranslatedContent != "" {
		return q.TranslatedContent
	}
	if config.Get().Language == config.EN && (q.Content == "" || strings.Contains(
		q.Content,
		"English description is not available for the problem.",
	)) {
		return q.TranslatedContent
	}
	return q.Content
}

func (q *QuestionData) GetFormattedContent() string {
	// TODO 处理上标、下标
	content := q.GetContent()
	converter := md.NewConverter("", true, nil)
	converter.Use(plugin.GitHubFlavored())
	content, err := converter.ConvertString(content)
	if err != nil {
		return content
	}
	content = text.WrapText(content, 100)
	content = utils.RemoveEmptyLine(content)
	return content
}

// GetExampleOutput parses example output from content and translatedContent
func (q *QuestionData) GetExampleOutput() []string {
	return nil
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

func QuestionBySlug(slug string, c Client) (*QuestionData, error) {
	q, err := c.GetQuestionData(slug)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func QuestionById(id string, c Client) (*QuestionData, error) {
	q := GetCache().GetById(id)
	if q != nil {
		return QuestionBySlug(q.Slug, c)
	}
	return nil, errors.New("no such question")
}

func Question(s string, c Client) (*QuestionData, error) {
	if s == "today" {
		return c.GetTodayQuestion()
	}
	q := GetCache().GetById(s)
	if q != nil {
		return QuestionBySlug(q.Slug, c)
	}
	return QuestionBySlug(s, c)
}
