package leetcode

import (
	"errors"
	"time"
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

func (ct *Contest) GetQuestionByNumber(num int) (*QuestionData, error) {
	if !ct.HasStarted() {
		return nil, ErrContestNotStarted
	}
	err := ct.Refresh()
	if err != nil {
		return nil, err
	}
	if len(ct.Questions) == 0 {
		return nil, errors.New("no questions in contest")
	}
	if num < 1 || num > len(ct.Questions) {
		return nil, errors.New("invalid question number")
	}

	q := ct.Questions[num-1]
	err = q.Fulfill()
	return q, err
}

func (ct *Contest) GetAllQuestions() ([]*QuestionData, error) {
	if !ct.HasStarted() {
		return nil, ErrContestNotStarted
	}
	err := ct.Refresh()
	if err != nil {
		return nil, err
	}
	if len(ct.Questions) == 0 {
		return nil, errors.New("no questions in contest")
	}

	for _, q := range ct.Questions {
		err := q.Fulfill()
		if err != nil {
			return nil, err
		}
	}
	return ct.Questions, nil
}

func (ct *Contest) Refresh() error {
	if len(ct.Questions) > 0 {
		return nil
	}
	contest, err := ct.client.GetContest(ct.TitleSlug)
	if err != nil {
		return err
	}
	*ct = *contest
	return nil
}
