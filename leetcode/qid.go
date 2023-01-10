package leetcode

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/j178/leetgo/config"
)

func QuestionFromCacheBySlug(slug string, c Client) (*QuestionData, error) {
	q := GetCache(c).GetBySlug(slug)
	if q != nil {
		q.client = c
		return q, nil
	}
	return nil, errors.New("no such question")
}

func QuestionFromCacheByID(id string, c Client) (*QuestionData, error) {
	q := GetCache(c).GetById(id)
	if q != nil {
		q.client = c
		return q, nil
	}
	return nil, errors.New("no such question")
}

// QuestionBySlug loads question data from cache first, if not found, fetch from leetcode.com
func QuestionBySlug(slug string, c Client) (*QuestionData, error) {
	q, err := QuestionFromCacheBySlug(slug, c)
	if err != nil {
		q, err = c.GetQuestionData(slug)
	}
	if q != nil {
		q.client = c
	}
	return q, err
}

func ParseQID(qid string, c Client) ([]*QuestionData, error) {
	var (
		q   *QuestionData
		qs  []*QuestionData
		err error
	)
	switch {
	case isNumber(qid):
		q, err = QuestionFromCacheByID(qid, c)
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
		_, qs, err = ParseContestQID(qid, c, true)
	}
	if err != nil {
		return nil, fmt.Errorf("invalid qid \"%s\": %w", qid, err)
	}
	// Try slug as last resort
	if q == nil && len(qs) == 0 {
		q, err = QuestionBySlug(qid, c)
		if err != nil {
			return nil, fmt.Errorf("invalid qid \"%s\": %w", qid, err)
		}
	}
	if q != nil {
		qs = []*QuestionData{q}
	}
	if len(qs) == 0 {
		return nil, fmt.Errorf("invalid qid \"%s\": no such question", qid)
	}
	return qs, nil
}

func ParseContestQID(qid string, c Client, withQuestions bool) (*Contest, []*QuestionData, error) {
	if len(qid) < 3 {
		return nil, nil, errors.New("invalid contest qid")
	}
	if strings.Count(qid, "/") != 1 {
		return nil, nil, errors.New("invalid contest qid")
	}

	var (
		contestSlug string
		questionNum = -1
		err         error
		q           *QuestionData
		qs          []*QuestionData
	)
	contestPat := regexp.MustCompile(`(?i)([wb])\D*(\d+)`)
	parts := strings.SplitN(qid, "/", 2)
	matches := contestPat.FindStringSubmatch(parts[0])
	if matches == nil {
		contestSlug = parts[0]
		if contestSlug == "last" {
			state := config.LoadState()
			if state.LastContest == "" {
				return nil, nil, errors.New("no last generated contest")
			}
			contestSlug = state.LastContest
		}
	} else {
		if matches[1][0] == 'w' || matches[1][0] == 'W' {
			contestSlug = "weekly-contest-" + matches[2]
		} else {
			contestSlug = "biweekly-contest-" + matches[2]
		}
	}
	if len(parts[1]) > 0 {
		questionNum, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, nil, fmt.Errorf("%s is not a number", parts[1])
		}
	}
	contest, err := c.GetContest(contestSlug)
	if err != nil {
		return nil, nil, fmt.Errorf("contest not found %s: %w", contestSlug, err)
	}

	if withQuestions {
		if questionNum > 0 {
			q, err = contest.GetQuestionByNumber(questionNum)
		} else {
			qs, err = contest.GetAllQuestions()
		}
		if err != nil {
			questionName := "<all>"
			if questionNum > 0 {
				questionName = strconv.Itoa(questionNum)
			}
			return contest, nil, fmt.Errorf("get contest question failed %s: %w", questionName, err)
		}
		if q != nil {
			qs = []*QuestionData{q}
		}
	}

	return contest, qs, nil
}

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
