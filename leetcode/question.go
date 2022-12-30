package leetcode

import (
	"encoding/json"
	"errors"
	"strconv"

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
	m.Manual = v["manual"].(bool)
	for _, param := range v["params"].([]any) {
		p := param.(map[string]any)
		m.Params = append(
			m.Params, MetaDataParam{
				Name: p["name"].(string),
				Type: p["type"].(string),
			},
		)
	}
	m.Return = MetaDataReturn{
		Type: v["return"].(map[string]any)["type"].(string),
		Size: int(v["return"].(map[string]any)["size"].(float64)),
	}
	return nil
}

type QuestionData struct {
	client             Client
	TitleSlug          string        `json:"titleSlug"`
	QuestionId         string        `json:"questionId"`
	QuestionFrontendId string        `json:"questionFrontendId"`
	Title              string        `json:"title"`
	TranslatedTitle    string        `json:"translatedTitle"`
	Difficulty         string        `json:"difficulty"`
	TopicTags          []TopicTag    `json:"topicTags"`
	IsPaidOnly         bool          `json:"isPaidOnly"`
	Content            string        `json:"content"`
	TranslatedContent  string        `json:"translatedContent"`
	Stats              Stats         `json:"stats"`
	Hints              []string      `json:"hints"`
	SimilarQuestions   string        `json:"similarQuestions"`
	SampleTestCase     string        `json:"sampleTestCase"`
	ExampleTestcases   string        `json:"exampleTestcases"`
	MetaData           MetaData      `json:"metaData"`
	CodeSnippets       []CodeSnippet `json:"codeSnippets"`
}

func (q *QuestionData) Url() string {
	return q.client.BaseURI() + "problems/" + q.TitleSlug + "/"
}

func (q *QuestionData) GetTitle() string {
	if config.Get().CN && q.TranslatedTitle != "" {
		return q.TranslatedTitle
	}
	return q.Title
}

func (q *QuestionData) GetContent() string {
	if config.Get().CN && q.TranslatedContent != "" {
		return q.TranslatedContent
	}
	return q.Content
}

func (q *QuestionData) TagSlugs() []string {
	var slugs []string
	for _, tag := range q.TopicTags {
		slugs = append(slugs, tag.Slug)
	}
	return slugs
}

func QuestionBySlug(slug string, c Client) (QuestionData, error) {
	q, err := c.GetQuestionData(slug)
	if err != nil {
		return QuestionData{}, err
	}
	return q, nil
}

func QuestionById(id string, c Client) (QuestionData, error) {
	q := GetCache().GetById(id)
	if q != nil {
		return QuestionBySlug(q.Slug, c)
	}
	return QuestionData{}, errors.New("no such question")
}

func Question(s string, c Client) (QuestionData, error) {
	if s == "today" {
		return c.GetTodayQuestion()
	}
	q, err := QuestionBySlug(s, c)
	if err == nil {
		return q, nil
	}
	return QuestionById(s, c)
}
