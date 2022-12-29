package leetcode

import (
	"errors"
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
	Stats              string        `json:"stats"`
	Hints              []string      `json:"hints"`
	SimilarQuestions   string        `json:"similarQuestions"`
	SampleTestCase     string        `json:"sampleTestCase"`
	ExampleTestcases   string        `json:"exampleTestcases"`
	MetaData           string        `json:"metaData"`
	CodeSnippets       []CodeSnippet `json:"codeSnippets"`
}

func (q *QuestionData) Url() string {
	return q.client.BaseURI() + "problems/" + q.TitleSlug + "/"
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
