package leetcode

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/j178/leetgo/config"
)

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
		return QuestionBySlug(q.TitleSlug, c)
	}
	return nil, errors.New("no such question")
}

func ParseQID(qid string, c Client) ([]*QuestionData, error) {
	var (
		q   *QuestionData
		qs  []*QuestionData
		err error
	)
	switch {
	case isNumber(qid):
		q, err = QuestionById(qid, c)
	case qid == "last":
		state := config.LoadState()
		if state.LastQuestion.Slug != "" {
			q, err = QuestionBySlug(state.LastQuestion.Slug, c)
		} else {
			err = errors.New("no last generated question")
		}
	case qid == "today":
		q, err = c.GetTodayQuestion()
	case strings.Contains(qid, "/"):
		qs, err = parseContestQID(qid, c)
	default:
		q, err = QuestionBySlug(qid, c)
	}

	if err != nil {
		return nil, fmt.Errorf("invalid qid \"%s\": %w", qid, err)
	}
	if q != nil {
		qs = []*QuestionData{q}
	}
	if len(qs) == 0 {
		return nil, fmt.Errorf("invalid qid \"%s\": no such question", qid)
	}
	return qs, nil
}

func parseContestQID(qid string, c Client) ([]*QuestionData, error) {
	if len(qid) <= 3 {
		return nil, errors.New("invalid contest qid")
	}
	if strings.Count(qid, "/") != 1 {
		return nil, errors.New("invalid contest qid")
	}

	var (
		contestSlug string
		questionNum = -1
		err         error
		q           *QuestionData
		qs          []*QuestionData
	)
	contestPat := regexp.MustCompile("([wb])[^0-9]*([0-9]+)")
	parts := strings.SplitN(qid, "/", 2)
	matches := contestPat.FindStringSubmatch(parts[0])
	if matches == nil {
		contestSlug = parts[0]
		if contestSlug == "last" {
			state := config.LoadState()
			if state.LastContest == "" {
				return nil, errors.New("no last generated contest")
			}
			contestSlug = state.LastContest
		}
	} else {
		if matches[1][0] == 'w' {
			contestSlug = "weekly-contest-" + matches[2]
		} else {
			contestSlug = "biweekly-contest-" + matches[2]
		}
	}
	if len(parts[1]) > 0 {
		questionNum, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid qid %s: %s is not a number", qid, parts[1])
		}
	}
	contest, err := c.GetContest(contestSlug)
	if err != nil {
		return nil, fmt.Errorf("contest not found %s: %w", contestSlug, err)
	}
	if questionNum > 0 {
		q, err = contest.GetQuestionByNumber(questionNum, c)
	} else {
		qs, err = contest.GetAllQuestions(c)
	}
	if err != nil {
		questionName := "<all>"
		if questionNum > 0 {
			questionName = strconv.Itoa(questionNum)
		}
		return nil, fmt.Errorf("get contest question failed %s: %w", questionName, err)
	}

	if q != nil {
		qs = []*QuestionData{q}
	}
	return qs, nil
}

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
