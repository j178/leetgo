package leetcode

import (
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/tidwall/gjson"

	"github.com/j178/leetgo/config"
)

type usClient struct {
	cnClient
}

func (c *usClient) BaseURI() string {
	return string(config.LeetCodeUS) + "/"
}

func (c *usClient) Login(username, password string) (*http.Response, error) {
	return nil, errors.New("leetcode.com does not support login with username and password")
}

func (c *usClient) GetQuestionData(slug string) (*QuestionData, error) {
	query := `
	query questionData($titleSlug: String!) {
		question(titleSlug: $titleSlug) {
			questionId
			questionFrontendId
			categoryTitle
			title
			titleSlug
			content
			isPaidOnly
			translatedTitle
			translatedContent
			difficulty
			status
			stats
			hints
			similarQuestions
			sampleTestCase
			exampleTestcases
			exampleTestcaseList
			metaData
			codeSnippets {
				lang
				langSlug
				code
			}
			topicTags {
				name
				slug
				translatedName
			}
		}
	}`
	q, err := c.getQuestionData(slug, query)
	if err != nil {
		return q, err
	}
	q.client = c
	return q, nil
}

func (c *usClient) GetAllQuestions() ([]*QuestionData, error) {
	var resp struct {
		UserName        string `json:"user_name"`
		NumSolved       int    `json:"num_solved"`
		NumTotal        int    `json:"num_total"`
		AcEasy          int    `json:"ac_easy"`
		AcMedium        int    `json:"ac_medium"`
		AcHard          int    `json:"ac_hard"`
		StatStatusPairs []struct {
			Stat struct {
				QuestionID         int    `json:"question_id"`
				QuestionFrontendID int    `json:"frontend_question_id"`
				QuestionTitle      string `json:"question__title"`
				QuestionTitleSlug  string `json:"question__title_slug"`
			} `json:"stat"`
			Status     string `json:"status"`
			Difficulty struct {
				Level int `json:"level"`
			} `json:"difficulty"`
			PaidOnly bool `json:"paid_only"`
		} `json:"stat_status_pairs"`
	}
	_, err := c.http.New().Get(problemsAllPath).ReceiveSuccess(&resp)
	if err != nil {
		return nil, err
	}
	qs := make([]*QuestionData, 0, len(resp.StatStatusPairs))
	for _, pair := range resp.StatStatusPairs {
		difficulty := ""
		switch pair.Difficulty.Level {
		case 1:
			difficulty = "Easy"
		case 2:
			difficulty = "Medium"
		case 3:
			difficulty = "Hard"
		}
		q := &QuestionData{
			QuestionId:         strconv.Itoa(pair.Stat.QuestionID),
			QuestionFrontendId: strconv.Itoa(pair.Stat.QuestionFrontendID),
			Title:              pair.Stat.QuestionTitle,
			TitleSlug:          pair.Stat.QuestionTitleSlug,
			IsPaidOnly:         pair.PaidOnly,
			Status:             pair.Status,
			Difficulty:         difficulty,
		}
		qs = append(qs, q)
	}
	return qs, nil
}

func (c *usClient) GetTodayQuestion() (*QuestionData, error) {
	query := `
	query questionOfToday {
		activeDailyCodingChallengeQuestion {
			question {
				titleSlug
			}
		}
	}`
	var resp gjson.Result
	_, err := c.graphqlPost(
		graphqlRequest{query: query}, &resp, nil,
	)
	if err != nil {
		return nil, err
	}
	slug := resp.Get("data.activeDailyCodingChallengeQuestion.question.titleSlug").Str
	return c.GetQuestionData(slug)
}

func (c *usClient) GetContest(contestSlug string) (*Contest, error) {
	ct, err := c.getContest(contestSlug)
	if err != nil {
		return nil, err
	}
	ct.client = c
	for i := range ct.Questions {
		ct.Questions[i].client = c
	}
	return ct, nil
}

func (c *usClient) GetContestQuestionData(contestSlug string, questionSlug string) (*QuestionData, error) {
	q, err := c.getContestQuestionData(contestSlug, questionSlug)
	if err != nil {
		return nil, err
	}
	q.client = c
	return q, nil
}

func (c *usClient) GetUpcomingContests() ([]*Contest, error) {
	// ContestNode does not have `registered` field
	// We have to call `/contest/api/info/slug` to get that info.
	query := `
{
    upcomingContests {
        title
        titleSlug
        description
        duration
        startTime
        originStartTime
        isVirtual
        containsPremium
		__typename
    }
}
`
	var resp gjson.Result
	_, err := c.graphqlPost(
		graphqlRequest{query: query}, &resp, nil,
	)
	if err != nil {
		return nil, err
	}
	var contests []*Contest
	for _, contestInfo := range resp.Get("data.upcomingContests").Array() {
		slug := contestInfo.Get("titleSlug").Str
		ct, err := c.GetContest(slug)
		var registered bool
		if err == nil {
			registered = ct.Registered
		}

		contests = append(
			contests, &Contest{
				client:          c,
				Id:              int(contestInfo.Get("id").Int()),
				TitleSlug:       slug,
				Title:           contestInfo.Get("title").Str,
				StartTime:       contestInfo.Get("startTime").Int(),
				OriginStartTime: contestInfo.Get("originStartTime").Int(),
				Duration:        int(contestInfo.Get("duration").Int()),
				IsVirtual:       contestInfo.Get("isVirtual").Bool(),
				Description:     contestInfo.Get("description").Str,
				Registered:      registered,
			},
		)
	}
	sort.Slice(
		contests, func(i, j int) bool {
			return contests[i].StartTime < contests[j].StartTime
		},
	)
	return contests, nil
}

// Cannot find a equivalent API for leetcode.com, use leetcode.cn instead.
func (c *usClient) GetQuestionTags() ([]QuestionTag, error) {
	return c.cnClient.GetQuestionTags()
}
