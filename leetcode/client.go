package leetcode

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dghubble/sling"
	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/utils"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/tidwall/gjson"
)

var (
	ErrPaidOnlyQuestion = errors.New("this is paid only question, you need to subscribe to LeetCode Premium")
	ErrTooManyRequests  = errors.New("you have submitted too frequently, please submit again later")
)

type Client interface {
	BaseURI() string
	Login(username, password string) (*http.Response, error)
	GetUserStatus() (*UserStatus, error)
	GetQuestionData(slug string) (*QuestionData, error)
	GetContest(contestSlug string) (*Contest, error)
	GetContestQuestionData(contestSlug, slug string) (*QuestionData, error)
	GetAllQuestions() ([]*QuestionData, error)
	GetTodayQuestion() (*QuestionData, error)
	InterpretSolution(q *QuestionData, lang string, code string, dataInput string) (
		*InterpretSolutionResult,
		error,
	)
	CheckResult(interpretId string) (CheckResult, error)
	Submit(q *QuestionData, lang string, code string) (string, error)
}

type cnClient struct {
	opt  Options
	http *sling.Sling
}

type Options struct {
	debug bool
	cred  CredentialsProvider
}

type ClientOption func(*Options)

func WithCredentials(cred CredentialsProvider) ClientOption {
	return func(o *Options) {
		o.cred = cred
	}
}

func NewClient(options ...ClientOption) Client {
	var opts Options
	for _, f := range options {
		f(&opts)
	}
	opts.debug = config.Debug

	httpClient := sling.New()
	httpClient.Add(
		"User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36",
	)
	httpClient.Add("Accept-Encoding", "gzip, deflate")
	httpClient.Add("x-requested-with", "XMLHttpRequest")
	httpClient.ResponseDecoder(
		smartDecoder{
			Debug:       opts.debug,
			LogResponse: true,
			LogLimit:    10 * 1024,
		},
	)

	cfg := config.Get()
	if cfg.LeetCode.Site == config.LeetCodeCN {
		c := &cnClient{
			http: httpClient,
			opt:  opts,
		}
		c.http.Base(c.BaseURI())
		c.http.Add("Referer", c.BaseURI())
		c.http.Add("Origin", string(config.LeetCodeCN))

		if cred, ok := opts.cred.(NeedClient); ok {
			cred.SetClient(c)
		}

		return c
	} else {
		c := &usClient{
			cnClient{
				http: httpClient,
				opt:  opts,
			},
		}
		c.http.Base(c.BaseURI())
		c.http.Add("Referer", c.BaseURI())
		c.http.Add("Origin", string(config.LeetCodeUS))

		if cred, ok := opts.cred.(NeedClient); ok {
			cred.SetClient(c)
		}

		return c
	}
}

type variables map[string]string

type graphqlRequest struct {
	path          string
	query         string
	operationName string
	variables     variables
}

const (
	graphQLPath = "/graphql"
	nojGoPath   = "/graphql/noj-go"
)

type defaultErrorHandler struct {
	msg string
}

func (c *cnClient) send(req *http.Request, result any, failure any) (*http.Response, error) {
	if c.opt.cred != nil {
		err := c.opt.cred.AddCredentials(req)
		if err != nil {
			return nil, err
		}
	}
	if c.opt.debug {
		bodyStr := []byte("<empty>")
		if req.Body != nil {
			bodyStr, _ = io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewReader(bodyStr))
		}
		hclog.L().Trace("request", "method", req.Method, "url", req.URL.String(), "body", utils.BytesToString(bodyStr))
	}

	if failure == nil {
		failure = &defaultErrorHandler{"<default>"}
	}

	// default error detection
	resp, err := c.http.Do(req, result, failure)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return resp, ErrTooManyRequests
	}
	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	if e, ok := failure.(*defaultErrorHandler); ok && e.msg != "<default>" {
		return resp, fmt.Errorf("request failed: %s", e.msg)
	}
	return nil, err
}

//nolint:unused
func (c *cnClient) graphqlGet(req graphqlRequest, result any, failure any) (*http.Response, error) {
	r, err := c.http.New().Get(req.path).QueryStruct(
		map[string]any{
			"query":         req.query,
			"operationName": req.operationName,
			"variables":     req.variables,
		},
	).Request()
	if err != nil {
		return nil, err
	}
	return c.send(r, result, failure)
}

func (c *cnClient) graphqlPost(req graphqlRequest, result any, failure any) (*http.Response, error) {
	r, err := c.http.New().Post(req.path).BodyJSON(
		map[string]any{
			"query":         req.query,
			"operationName": req.operationName,
			"variables":     req.variables,
		},
	).Request()
	if err != nil {
		return nil, err
	}
	return c.send(r, result, failure)
}

func (c *cnClient) jsonGet(url string, query any, result any, failure any) (*http.Response, error) {
	r, err := c.http.New().Get(url).QueryStruct(query).Request()
	if err != nil {
		return nil, err
	}
	return c.send(r, result, failure)
}

func (c *cnClient) jsonPost(url string, json any, result any, failure any) (*http.Response, error) {
	r, err := c.http.New().Post(url).BodyJSON(json).Request()
	if err != nil {
		return nil, err
	}
	return c.send(r, result, failure)
}

func (c *cnClient) BaseURI() string {
	return string(config.LeetCodeCN) + "/"
}

func (c *cnClient) Login(username, password string) (*http.Response, error) {
	// touch "csrftoken" cookie
	req, _ := c.http.New().Post(graphQLPath).BodyJSON(
		map[string]any{
			"query":         `query nojGlobalData {\n  siteRegion\n  chinaHost\n  websocketUrl\n}`,
			"operationName": "nojGlobalData",
			"variables":     nil,
		},
	).Request()
	resp, err := c.http.Do(req, nil, nil)
	if err != nil {
		return resp, err
	}

	var csrfToken string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrftoken" {
			csrfToken = cookie.Value
			break
		}
	}
	if csrfToken == "" {
		return nil, errors.New("csrf token not found")
	}

	cc := c.http.New()
	// Disable redirect
	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	cc.Client(httpClient)

	body := struct {
		Login               string `url:"login"`
		Password            string `url:"password"`
		CsrfMiddlewareToken string `url:"csrfmiddlewaretoken"`
	}{username, password, csrfToken}
	req, err = cc.Post("/accounts/login/").BodyForm(body).Request()
	if err != nil {
		return nil, err
	}
	resp, err = cc.Do(req, nil, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusBadRequest {
		return nil, errors.New("login failed, please check your username and password")
	}
	return resp, nil
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
	_, err := c.graphqlPost(
		graphqlRequest{
			path:          nojGoPath,
			query:         query,
			operationName: "userStatusGlobal",
			variables:     nil,
		}, &resp, nil,
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
	_, err := c.graphqlPost(
		graphqlRequest{
			path:          graphQLPath,
			query:         query,
			operationName: "questionData",
			variables:     variables{"titleSlug": slug},
		}, &resp, nil,
	)
	if err != nil {
		return nil, err
	}
	q := resp.Data.Question
	if q.TitleSlug == "" {
		return nil, errors.New("question not found")
	}
	if q.IsPaidOnly && q.Content == "" {
		return nil, ErrPaidOnlyQuestion
	}
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
	_, err := c.graphqlPost(
		graphqlRequest{
			path:          graphQLPath,
			query:         query,
			operationName: "AllQuestionUrls",
			variables:     nil,
		}, &resp, nil,
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
	dec := progressDecoder{smartDecoder{LogResponse: false}, tracker}
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
	_, err := c.graphqlPost(
		graphqlRequest{
			path:          graphQLPath,
			query:         query,
			operationName: "questionOfToday",
			variables:     nil,
		}, &resp, nil,
	)
	if err != nil {
		return nil, err
	}
	slug := resp.Get("data.todayRecord.0.question.titleSlug").Str
	return c.GetQuestionData(slug)
}

func (c *cnClient) GetContest(contestSlug string) (*Contest, error) {
	return &Contest{client: c, TitleSlug: contestSlug}, nil
}

func (c *cnClient) GetContestQuestionData(contestSlug, slug string) (*QuestionData, error) {
	return nil, nil
}

// 每次 "运行代码" 会产生两个 submission, 一个是运行我们的代码，一个是运行标程。

func (c *cnClient) InterpretSolution(q *QuestionData, lang string, code string, dataInput string) (
	*InterpretSolutionResult,
	error,
) {
	url := fmt.Sprintf("%sproblems/%s/interpret_solution/", c.BaseURI(), q.TitleSlug)
	var resp InterpretSolutionResult
	_, err := c.jsonPost(
		url, map[string]any{
			"lang":        lang,
			"question_id": q.QuestionId,
			"typed_code":  code,
			"data_input":  dataInput,
		}, &resp, nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp, err
}

func (c *cnClient) CheckResult(submissionId string) (
	CheckResult,
	error,
) {
	url := fmt.Sprintf("%s/submissions/detail/%s/check/", c.BaseURI(), submissionId)
	var result gjson.Result
	_, err := c.jsonGet(url, nil, &result, nil)
	if err != nil {
		return nil, err
	}
	if result.Get("question_id").Exists() {
		var r SubmitCheckResult
		err = json.Unmarshal(utils.StringToBytes(result.Raw), &r)
		return &r, err
	}
	var r RunCheckResult
	err = json.Unmarshal(utils.StringToBytes(result.Raw), &r)
	return &r, err
}

func (c *cnClient) Submit(q *QuestionData, lang string, code string) (string, error) {
	url := fmt.Sprintf("%sproblems/%s/submit/", c.BaseURI(), q.TitleSlug)
	var resp gjson.Result
	_, err := c.jsonPost(
		url, map[string]any{
			"lang":         lang,
			"questionSlug": q.TitleSlug,
			"question_id":  q.QuestionId,
			"typed_code":   code,
		}, &resp, nil,
	)
	return resp.Get("submission_id").String(), err
}
