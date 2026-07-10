package leetcode

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/j178/leetgo/config"
)

func questionFromSavedState(c Client, id string) (*QuestionData, bool) {
	state := config.LoadState()
	if saved, ok := state.Questions[id]; ok {
		q := questionFromSaved(c, saved)
		return q, q != nil
	}
	last := state.LastQuestion
	if last.Slug == "" || len(last.MetaData) == 0 {
		return nil, false
	}
	if id != "last" && id != last.Slug && id != last.FrontendID {
		return nil, false
	}
	q := questionFromSaved(c, last)
	return q, q != nil
}

func questionFromSaved(c Client, saved config.LastQuestion) *QuestionData {
	var meta MetaData
	if err := json.Unmarshal(saved.MetaData, &meta); err != nil {
		return nil
	}
	q := &QuestionData{
		TitleSlug:          saved.Slug,
		QuestionFrontendId: saved.FrontendID,
		Content:            saved.Content,
		TranslatedContent:  saved.TranslatedContent,
		MetaData:           meta,
	}
	if saved.ContestSlug != "" {
		q.contest = &Contest{TitleSlug: saved.ContestSlug}
		q.contest.client = c
	}
	q.client = c
	return q
}

func contestFromSavedState(c Client, contestSlug string, questionNum int, withQuestions bool) (*Contest, []*QuestionData, bool) {
	state := config.LoadState()
	savedContest, ok := state.Contests[contestSlug]
	if !ok || len(savedContest.Questions) == 0 {
		return nil, nil, false
	}

	contest := &Contest{
		client:    c,
		TitleSlug: contestSlug,
		Questions: make([]*QuestionData, 0, len(savedContest.Questions)),
	}
	if !withQuestions {
		return contest, nil, true
	}

	if questionNum > 0 {
		if questionNum > len(savedContest.Questions) {
			return nil, nil, false
		}
		q, ok := questionFromSavedState(c, savedContest.Questions[questionNum-1])
		if !ok {
			return nil, nil, false
		}
		q.contest = contest
		contest.Questions = []*QuestionData{q}
		return contest, contest.Questions, true
	}

	for _, slug := range savedContest.Questions {
		q, ok := questionFromSavedState(c, slug)
		if !ok {
			return nil, nil, false
		}
		q.contest = contest
		contest.Questions = append(contest.Questions, q)
	}
	return contest, contest.Questions, true
}

func QuestionFromCacheBySlug(slug string, c Client) (*QuestionData, error) {
	if q, ok := questionFromSavedState(c, slug); ok {
		return q, nil
	}
	q := GetCache(c).GetBySlug(slug)
	if q != nil {
		q.client = c
		return q, nil
	}
	return nil, ErrQuestionNotFound
}

func QuestionFromCacheByID(id string, c Client) (*QuestionData, error) {
	if q, ok := questionFromSavedState(c, id); ok {
		return q, nil
	}
	q := GetCache(c).GetById(id)
	if q != nil {
		q.client = c
		return q, nil
	}
	return nil, ErrQuestionNotFound
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
			err = errors.New("invalid qid: last generated question not found")
		}
	case qid == "today":
		q, err = c.GetTodayQuestion()
	case qid == "yesterday":
		q, err = c.GetQuestionOfDate(time.Now().AddDate(0, 0, -1))
	case strings.HasPrefix(qid, "today-"):
		var n int
		n, err = strconv.Atoi(qid[6:])
		if err == nil {
			q, err = c.GetQuestionOfDate(time.Now().AddDate(0, 0, -n))
		}
	case strings.Contains(qid, "/"):
		_, qs, err = ParseContestQID(qid, c, true)
		if err != nil {
			return nil, err
		}
	}
	if errors.Is(err, ErrQuestionNotFound) {
		err = nil
	}
	if err != nil {
		return nil, fmt.Errorf("invalid qid: %w", err)
	}
	if q == nil && len(qs) == 0 {
		q, err = QuestionBySlug(qid, c)
		if errors.Is(err, ErrQuestionNotFound) {
			q, err = QuestionFromCacheByID(qid, c)
		}
		if err != nil {
			return nil, fmt.Errorf("invalid qid: %w", err)
		}
	}
	if q != nil {
		qs = []*QuestionData{q}
	}
	if len(qs) == 0 {
		return nil, errors.New("invalid qid: no such question")
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
				return nil, nil, errors.New("invalid contest qid: last contest not found")
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
			return nil, nil, fmt.Errorf("invalid contest qid: %s is not a number", parts[1])
		}
	}

	if contest, qs, ok := contestFromSavedState(c, contestSlug, questionNum, withQuestions); ok {
		return contest, qs, nil
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
			return contest, nil, fmt.Errorf("get contest question %s failed: %w", questionName, err)
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
