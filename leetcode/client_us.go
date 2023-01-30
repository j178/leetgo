package leetcode

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/goccy/go-json"
	"github.com/tidwall/gjson"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/utils"
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

// GetUserStatus can be reused from cnClient

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
	// TODO implement me
	panic("implement me")
}

func (c *usClient) GetTodayQuestion() (*QuestionData, error) {
	// TODO implement me
	panic("implement me")
}

func (c *usClient) GetUpcomingContests() ([]*Contest, error) {
	query := `
{
    contestUpcomingContests {
        containsPremium
        title
        titleSlug
        description
        startTime
        duration
        originStartTime
        isVirtual
        registered
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
	for _, contestInfo := range resp.Get("data.contestUpcomingContests").Array() {
		contests = append(
			contests, &Contest{
				client:          c,
				Id:              int(contestInfo.Get("id").Int()),
				TitleSlug:       contestInfo.Get("titleSlug").Str,
				Title:           contestInfo.Get("title").Str,
				StartTime:       contestInfo.Get("startTime").Int(),
				OriginStartTime: contestInfo.Get("originStartTime").Int(),
				Duration:        int(contestInfo.Get("duration").Int()),
				IsVirtual:       contestInfo.Get("isVirtual").Bool(),
				Description:     contestInfo.Get("description").Str,
				Registered:      contestInfo.Get("registered").Bool(),
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

func (c *usClient) RegisterContest(slug string) error {
	path := fmt.Sprintf(contestRegisterPath, slug)
	_, err := c.jsonPost(path, nil, nil, nil)
	if e, ok := err.(unexpectedStatusCode); ok && e.Code == http.StatusFound {
		err = nil
	}
	return err
}

func (c *usClient) UnregisterContest(slug string) error {
	path := fmt.Sprintf(contestRegisterPath, slug)
	req, _ := c.http.New().Delete(path).Request()
	_, err := c.send(req, nil, nil)
	return err
}

func (c *usClient) GetQuestionsByFilter(f QuestionFilter, limit int, skip int) (QuestionList, error) {
	query := `
query problemsetQuestionList($categorySlug: String, $limit: Int, $skip: Int, $filters: QuestionListFilterInput) {
  problemsetQuestionList(
    categorySlug: $categorySlug
    limit: $limit
    skip: $skip
    filters: $filters
  ) {
    hasMore
    total
    questions {
      difficulty
      frontendQuestionId
      status
      title
      titleCn
      titleSlug
      topicTags {
        name
        nameTranslated
        id
        slug
      }
    }
  }
}
`
	vars := map[string]any{
		"categorySlug": "algorithms",
		"limit":        limit,
		"skip":         skip,
		"filters":      f,
	}
	var resp gjson.Result
	_, err := c.graphqlPost(
		graphqlRequest{
			query:     query,
			variables: vars,
		}, &resp, nil,
	)
	if err != nil {
		return QuestionList{}, err
	}

	var result QuestionList
	questionList := resp.Get("data.problemsetQuestionList")
	err = json.Unmarshal(utils.StringToBytes(questionList.Raw), &result)
	if err != nil {
		return QuestionList{}, err
	}
	for _, q := range result.Questions {
		q.client = c
		q.partial = 1
	}

	return result, err
}

func (c *usClient) GetQuestionTags() ([]QuestionTag, error) {
	query := `
query questionTagTypeWithTags {
    questionTagTypeWithTags {
        name
        transName
        tagRelation {
            questionNum
            tag {
                name
                id
                nameTranslated
                slug
            }
        }
    }
}
`
	var resp gjson.Result
	_, err := c.graphqlPost(graphqlRequest{query: query}, &resp, nil)
	if err != nil {
		return nil, err
	}
	var tags []QuestionTag
	for _, tagType := range resp.Get("data.questionTagTypeWithTags").Array() {
		tagTypeName := tagType.Get("name").Str
		tagTypeTransName := tagType.Get("transName").Str
		for _, tagInfo := range tagType.Get("tagRelation").Array() {
			tag := QuestionTag{
				TypeName:       tagTypeName,
				TypeTransName:  tagTypeTransName,
				Id:             tagInfo.Get("tag.id").Str,
				Name:           tagInfo.Get("tag.name").Str,
				NameTranslated: tagInfo.Get("tag.nameTranslated").Str,
				Slug:           tagInfo.Get("tag.slug").Str,
			}
			tags = append(tags, tag)
		}
	}

	return tags, nil
}
