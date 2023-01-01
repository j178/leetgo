package leetcode

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dghubble/sling"
	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/tidwall/gjson"
)

type Client interface {
	BaseURI() string
	WithCredentials(provider CredentialsProvider) Client
	Login(username, password string) (*http.Response, error)
	GetUserStatus() (*UserStatus, error)
	GetQuestionData(slug string) (*QuestionData, error)
	GetAllQuestions() ([]*QuestionData, error)
	GetTodayQuestion() (*QuestionData, error)
}

type Variables map[string]string

type cnClient struct {
	cred CredentialsProvider
	http *sling.Sling
}

func NewClient() Client {
	httpClient := sling.New()
	httpClient.Add(
		"User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36",
	)
	httpClient.Add("Accept-Encoding", "gzip, deflate, br")
	httpClient.Add("x-requested-with", "XMLHttpRequest")
	httpClient.ResponseDecoder(smartDecoder{LogResponseData: true})

	cfg := config.Get()
	if cfg.LeetCode.Site == config.LeetCodeCN {
		c := &cnClient{
			http: httpClient,
		}
		c.http.Base(c.BaseURI())
		c.http.Add("Referer", c.BaseURI())
		c.http.Add("Origin", string(config.LeetCodeCN))
		return c
	} else {
		panic(fmt.Sprintf("site not supported yet: %s", cfg.LeetCode.Site))
	}
}

func (c *cnClient) WithCredentials(provider CredentialsProvider) Client {
	cc := &cnClient{
		cred: provider,
		http: c.http.New(),
	}
	return cc
}

type request struct {
	path          string
	query         string
	operationName string
	variables     Variables
	needAuth      bool
}

const (
	graphQLPath = "/graphql"
	nojGoPath   = "/graphql/noj-go"
)

//nolint:unused
func (c *cnClient) graphqlGet(req request, result any) error {
	r, err := c.http.New().Get(req.path).QueryStruct(
		map[string]any{
			"query":         req.query,
			"operationName": req.operationName,
			"variables":     req.variables,
		},
	).Request()
	if err != nil {
		return err
	}
	if req.needAuth && c.cred == nil {
		return errors.New("no credentials provider set")
	}
	if req.needAuth {
		err = c.cred.AddCredentials(r, c)
		if err != nil {
			return err
		}
	}
	hclog.L().Trace("request", "method", "GET", "url", r.URL.String())
	_, err = c.http.Do(r, result, nil)
	return err
}

func (c *cnClient) graphqlPost(req request, result any) error {
	r, err := c.http.New().Post(req.path).BodyJSON(
		map[string]any{
			"query":         req.query,
			"operationName": req.operationName,
			"variables":     req.variables,
		},
	).Request()
	if err != nil {
		return err
	}
	if req.needAuth && c.cred == nil {
		return errors.New("no credentials provider set")
	}
	if req.needAuth {
		err = c.cred.AddCredentials(r, c)
		if err != nil {
			return err
		}
	}
	hclog.L().Trace("request", "method", "POST", "url", r.URL.String())
	_, err = c.http.Do(r, result, nil)
	return err
}

func (c *cnClient) BaseURI() string {
	return string(config.LeetCodeCN) + "/"
}

func (c *cnClient) Login(username, password string) (*http.Response, error) {
	return nil, errors.New("not implemented")
}

func (c *cnClient) GetUserStatus() (*UserStatus, error) {
	query := `
query userStatusGlobal {
  userStatus {
    isSignedIn
    username
    realName
    userSlug
    avatar
    activeSessionId
	isPremium
  }
}`
	var resp struct {
		Data struct {
			UserStatus UserStatus `json:"userStatus"`
		} `json:"data"`
	}
	err := c.graphqlPost(
		request{
			path:          nojGoPath,
			query:         query,
			operationName: "userStatusGlobal",
			variables:     nil,
			needAuth:      true,
		}, &resp,
	)
	if err != nil {
		return nil, err
	}
	userStatus := resp.Data.UserStatus
	return &userStatus, nil
}

func (c *cnClient) GetQuestionData(slug string) (*QuestionData, error) {
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
			jsonExampleTestcases
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

	var resp struct {
		Data struct {
			Question QuestionData `json:"question"`
		}
	}
	err := c.graphqlPost(
		request{
			path:          graphQLPath,
			query:         query,
			operationName: "questionData",
			variables:     Variables{"titleSlug": slug},
			needAuth:      false,
		}, &resp,
	)
	if err != nil {
		return nil, err
	}
	if resp.Data.Question.TitleSlug == "" {
		return nil, errors.New("question not found")
	}
	q := resp.Data.Question
	q.client = c
	return &q, nil
}

func (c *cnClient) GetAllQuestions() ([]*QuestionData, error) {
	query := `
	query AllQuestionUrls {
		allQuestionUrls {
			questionUrl
		}
	}
	`
	var resp gjson.Result
	err := c.graphqlPost(
		request{
			path:          graphQLPath,
			query:         query,
			operationName: "AllQuestionUrls",
			variables:     nil,
			needAuth:      false,
		}, &resp,
	)
	if err != nil {
		return nil, err
	}
	url := resp.Get("data.allQuestionUrls.questionUrl").Str

	hclog.L().Trace("request", "url", url)
	tracker := &progress.Tracker{
		Message: "Downloading questions",
		Total:   0,
		Units:   progress.UnitsBytes,
	}
	pw := progress.NewWriter()
	pw.SetAutoStop(true)
	pw.AppendTracker(tracker)
	pw.SetStyle(progress.StyleBlocks)
	pw.Style().Visibility.ETA = false
	pw.Style().Visibility.ETAOverall = false

	go pw.Render()

	var qs []*QuestionData
	dec := progressDecoder{smartDecoder{LogResponseData: false}, tracker}
	_, err = c.http.New().Get(url).ResponseDecoder(dec).ReceiveSuccess(&qs)
	if err != nil {
		return nil, err
	}
	// Sleep a while to make sure the progress bar is rendered.
	time.Sleep(time.Millisecond * 100)
	return qs, err
}

func (c *cnClient) GetTodayQuestion() (*QuestionData, error) {
	query := `
    query questionOfToday {
        todayRecord {
            question {
                titleSlug
            }
        }
    }`
	var resp gjson.Result
	err := c.graphqlPost(
		request{
			path:          graphQLPath,
			query:         query,
			operationName: "questionOfToday",
			variables:     nil,
			needAuth:      false,
		}, &resp,
	)
	if err != nil {
		return nil, err
	}
	slug := resp.Get("data.todayRecord.0.question.titleSlug").Str
	return c.GetQuestionData(slug)
}
