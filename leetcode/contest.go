package leetcode

import "github.com/hashicorp/go-hclog"

type Contest struct {
	client    Client
	TitleSlug string `json:"titleSlug"`
	Title     string `json:"title"`
	StartTime string `json:"startTime"`
	Questions []*QuestionData
}

func (ct *Contest) GetQuestion(slug string) (*QuestionData, error) {
	for _, q := range ct.Questions {
		if q.TitleSlug == slug {
			return q, nil
		}
	}
	return nil, nil
}

func (ct *Contest) GetQuestionByNumber(num int, c Client) (*QuestionData, error) {
	hclog.L().Info("get question by number", "contest", ct.TitleSlug, "num", num)
	return nil, nil
}

func (ct *Contest) GetAllQuestions(c Client) ([]*QuestionData, error) {
	hclog.L().Info("get all questions", "contest", ct.TitleSlug)
	return nil, nil
}
