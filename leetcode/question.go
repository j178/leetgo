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
	var v map[string]any
	unquoted, err := strconv.Unquote(utils.BytesToString(data))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(utils.StringToBytes(unquoted), &v); err != nil {
		return err
	}
	s.TotalAccepted = v["totalAccepted"].(string)
	s.TotalSubmission = v["totalSubmission"].(string)
	s.TotalAcceptedRaw = int(v["totalAcceptedRaw"].(float64))
	s.TotalSubmissionRaw = int(v["totalSubmissionRaw"].(float64))
	s.ACRate = v["acRate"].(string)
	return nil
}

type MetaDataParam struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type MetaDataReturn struct {
	Type string `json:"type"`
	Size int    `json:"size"`
}

type MetaData struct {
	Name   string          `json:"name"`
	Params []MetaDataParam `json:"params"`
	Return MetaDataReturn  `json:"return"`
	Manual bool            `json:"manual"`
}

func (m *MetaData) UnmarshalJSON(data []byte) error {
	var v map[string]any
	unquoted, err := strconv.Unquote(utils.BytesToString(data))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(utils.StringToBytes(unquoted), &v); err != nil {
		return err
	}
	m.Name = v["name"].(string)
	if manual, ok := v["manual"].(bool); ok {
		m.Manual = manual
	}
	for _, param := range v["params"].([]any) {
		p := param.(map[string]any)
		m.Params = append(
			m.Params, MetaDataParam{
				Name: p["name"].(string),
				Type: p["type"].(string),
			},
		)
	}
	ret := v["return"].(map[string]any)
	m.Return.Type = ret["type"].(string)
	if size, ok := ret["size"].(float64); ok {
		m.Return.Size = int(size)
	}
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
	var v []map[string]string
	if err := json.Unmarshal(utils.StringToBytes(unquoted), &v); err != nil {
		return err
	}
	for _, q := range v {
		*s = append(
			*s, SimilarQuestion{
				Title:           q["title"],
				TitleSlug:       q["titleSlug"],
				Difficulty:      q["difficulty"],
				TranslatedTitle: q["translatedTitle"],
			},
		)
	}
	return nil
}

type QuestionData struct {
	client               Client
	contestSlug          string
	TitleSlug            string               `json:"titleSlug"`
	QuestionId           string               `json:"questionId"`
	QuestionFrontendId   string               `json:"questionFrontendId"`
	CategoryTitle        string               `json:"CategoryTitle"`
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
