package leetcode

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"rogchap.com/v8go"
)

var ErrContestNotStarted = errors.New("contest has not started")

type Contest struct {
	client          Client
	Id              int
	TitleSlug       string
	Title           string
	StartTime       int64
	OriginStartTime int64
	Duration        int
	Description     string
	Questions       []*QuestionData
	Registered      bool
	ContainsPremium bool
	IsVirtual       bool
}

func (ct *Contest) HasStarted() bool {
	return time.Unix(ct.StartTime, 0).Before(time.Now())
}

func (ct *Contest) HasFinished() bool {
	return time.Unix(ct.StartTime, 0).Add(time.Duration(ct.Duration) * time.Second).Before(time.Now())
}

func (ct *Contest) TimeTillStart() time.Duration {
	return time.Until(time.Unix(ct.StartTime, 0))
}

func (ct *Contest) checkAccessQuestions() error {
	if !ct.HasStarted() {
		return ErrContestNotStarted
	}
	if len(ct.Questions) > 0 {
		return nil
	}
	err := ct.Refresh()
	if err != nil {
		return err
	}
	if len(ct.Questions) == 0 {
		return errors.New("no questions in contest")
	}
	return nil
}

func (ct *Contest) GetQuestionNumber(slug string) (int, error) {
	err := ct.checkAccessQuestions()
	if err != nil {
		return 0, err
	}
	for i, q2 := range ct.Questions {
		if q2.TitleSlug == slug {
			return i + 1, nil
		}
	}
	return 0, errors.New("question not found")
}

func (ct *Contest) GetQuestionByNumber(num int) (*QuestionData, error) {
	err := ct.checkAccessQuestions()
	if err != nil {
		return nil, err
	}
	if num < 1 || num > len(ct.Questions) {
		return nil, errors.New("invalid question number")
	}

	q := ct.Questions[num-1]
	err = q.Fulfill()
	return q, err
}

func (ct *Contest) GetAllQuestions() ([]*QuestionData, error) {
	err := ct.checkAccessQuestions()
	if err != nil {
		return nil, err
	}
	return ct.Questions, nil
}

func (ct *Contest) Refresh() error {
	contest, err := ct.client.GetContest(ct.TitleSlug)
	if err != nil {
		return err
	}
	*ct = *contest
	return nil
}

func parseContestQuestionHtml(htmlText []byte) (*QuestionData, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlText))
	if err != nil {
		return nil, err
	}
	var scriptText string
	doc.Find("script").EachWithBreak(
		func(i int, selection *goquery.Selection) bool {
			node := selection.Get(0)
			if node.FirstChild != nil && strings.Contains(node.FirstChild.Data, "var pageData") {
				scriptText = node.FirstChild.Data
				return false
			}
			return true
		},
	)
	if scriptText == "" {
		return nil, errors.New("question data not found")
	}
	ctx := v8go.NewContext()
	val, err := ctx.RunScript(scriptText, "script.js")
	if err != nil {
		return nil, err
	}
	questionData, err := val.AsObject()
	if err != nil {
		return nil, err
	}
	fmt.Println(questionData)
	return nil, errors.New("not implemented")
}
