package leetcode

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/avast/retry-go"
	"github.com/charmbracelet/log"
	"github.com/dghubble/sling"
	"github.com/goccy/go-json"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/tidwall/gjson"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/utils"
)

var (
	ErrPaidOnlyQuestion  = errors.New("this is paid only question, you need to subscribe to LeetCode Premium")
	ErrQuestionNotFound  = errors.New("no such question")
	ErrContestNotStarted = errors.New("contest has not started")
)

type UnexpectedStatusCode struct {
	Code int
	Body string
}

func (e UnexpectedStatusCode) IsError() bool {
	return e.Code != 0
}

func (e UnexpectedStatusCode) Error() string {
	body := "<empty body>"
	if len(e.Body) > 500 {
		body = e.Body[:500] + "..."
	}
	return fmt.Sprintf("[%d %s] %s", e.Code, http.StatusText(e.Code), body)
}

func NewUnexpectedStatusCode(code int, body []byte) UnexpectedStatusCode {
	err := UnexpectedStatusCode{Code: code}
	switch code {
	case http.StatusTooManyRequests:
		err.Body = "LeetCode limited you access rate, you may be submitting too frequently"
	case http.StatusForbidden:
		err.Body = "Access is forbidden, your cookies may have expired or LeetCode has restricted its API access"
	default:
		err.Body = utils.BytesToString(body)
	}
	return err
}

type Client interface {
	BaseURI() string
	Inspect(typ string) (map[string]any, error)
	Login(username, password string) (*http.Response, error)
	GetUserStatus() (*UserStatus, error)
	GetQuestionData(slug string) (*QuestionData, error)
	GetAllQuestions() ([]*QuestionData, error)
	GetTodayQuestion() (*QuestionData, error)
	GetQuestionOfDate(date time.Time) (*QuestionData, error)
	GetQuestionsByFilter(f QuestionFilter, limit int, skip int) (QuestionList, error)
	GetQuestionTags() ([]QuestionTag, error)
	RunCode(q *QuestionData, lang string, code string, dataInput string) (
		*InterpretSolutionResult,
		error,
	)
	SubmitCode(q *QuestionData, lang string, code string) (string, error)
	CheckResult(interpretId string) (CheckResult, error)
	GetUpcomingContests() ([]*Contest, error)
	GetContest(contestSlug string) (*Contest, error)
	GetContestQuestionData(contestSlug string, questionSlug string) (*QuestionData, error)
	RegisterContest(slug string) error
	UnregisterContest(slug string) error
	GetStreakCounter() (StreakCounter, error)
}

type cnClient struct {
	opt  Options
	http *sling.Sling
}

type Options struct {
	debug bool
	cred  CredentialsProvider
}

func NewClient(cred CredentialsProvider) Client {
	opts := Options{
		cred:  cred,
		debug: config.Debug,
	}

	httpClient := sling.New()
	httpClient.Add(
		"User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
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
	httpClient.Client(
		&http.Client{
			CheckRedirect: nonFollowRedirect,
			Transport: &http.Transport{
				// Disable http2
				TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
			},
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

func nonFollowRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

type graphqlRequest struct {
	path          string
	query         string
	operationName string
	variables     map[string]any
	authType      authType
}

type authType int

const (
	withAuth authType = iota
	withoutAuth
	requireAuth
)

const (
	graphQLPath           = "/graphql"
	graphQLNoj            = "/graphql/noj-go/"
	accountLoginPath      = "/accounts/login/"
	contestInfoPath       = "/contest/api/info/%s/"
	contestProblemsPath   = "/contest/%s/problems/%s/"
	contestRunCodePath    = "/contest/api/%s/problems/%s/interpret_solution/"
	runCodePath           = "/problems/%s/interpret_solution/"
	contestSubmitCodePath = "/contest/api/%s/problems/%s/submit/"
	submitCodePath        = "/problems/%s/submit/"
	checkResultPath       = "/submissions/detail/%s/check/"
	contestRegisterPath   = "/contest/api/%s/register/"
	problemsAllPath       = "/api/problems/all/"
	problemsApiTagsPath   = "/problems/api/tags/"
)

func (c *cnClient) send(req *http.Request, authType authType, result any) (*http.Response, error) {
	switch authType {
	case withoutAuth:
	case withAuth:
		if err := c.opt.cred.AddCredentials(req); err != nil {
			log.Warn("add credentials failed, continue requesting without credentials", "err", err)
		}
	case requireAuth:
		if err := c.opt.cred.AddCredentials(req); err != nil {
			return nil, err
		}
	}

	if c.opt.debug {
		bodyStr := []byte("<empty>")
		if req.Body != nil {
			bodyStr, _ = io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewReader(bodyStr))
		}
		log.Debug("request", "method", req.Method, "url", req.URL.String(), "body", utils.BytesToString(bodyStr))
	}

	err := retry.Do(
		func() error {
			var (
				err     error
				respErr UnexpectedStatusCode
			)
			_, err = c.http.Do(req, result, &respErr)
			if err != nil {
				return err
			}
			if respErr.IsError() {
				return respErr
			}
			return nil
		},
		retry.RetryIf(
			func(err error) bool {
				// Do not retry on 429
				var e UnexpectedStatusCode
				if errors.As(err, &e) && e.Code == http.StatusTooManyRequests {
					return false
				}
				return true
			},
		),
		retry.Attempts(3),
		retry.LastErrorOnly(true),
		retry.OnRetry(
			func(n uint, err error) {
				log.Warn("retry", "url", req.URL.String(), "attempt", n, "error", err)
			},
		),
	)

	return nil, err
}

//nolint:unused
func (c *cnClient) graphqlGet(req graphqlRequest, result any) (*http.Response, error) {
	type params struct {
		Query         string `url:"query"`
		OperationName string `url:"operationName"`
		Variables     string `url:"variables"`
	}
	p := params{Query: req.query}
	if req.operationName != "" {
		p.OperationName = req.operationName
	}
	if req.variables != nil {
		v, _ := json.Marshal(req.variables)
		p.Variables = string(v)
	}
	path := graphQLPath
	if req.path != "" {
		path = req.path
	}
	r, err := c.http.New().Get(path).QueryStruct(p).Request()
	if err != nil {
		return nil, err
	}
	return c.send(r, req.authType, result)
}

func (c *cnClient) graphqlPost(req graphqlRequest, result any) (*http.Response, error) {
	v := req.variables
	if v == nil {
		v = make(map[string]any)
	}
	body := map[string]any{
		"query":         req.query,
		"operationName": req.operationName,
		"variables":     v,
	}
	path := graphQLPath
	if req.path != "" {
		path = req.path
	}
	r, err := c.http.New().Post(path).BodyJSON(body).Request()
	if err != nil {
		return nil, err
	}
	return c.send(r, req.authType, result)
}

func (c *cnClient) jsonGet(url string, query any, authType authType, result any) (*http.Response, error) {
	r, err := c.http.New().Get(url).QueryStruct(query).Request()
	if err != nil {
		return nil, err
	}
	return c.send(r, authType, result)
}

func (c *cnClient) jsonPost(url string, json any, authType authType, result any) (*http.Response, error) {
	r, err := c.http.New().Post(url).BodyJSON(json).Request()
	if err != nil {
		return nil, err
	}
	return c.send(r, authType, result)
}

func (c *cnClient) BaseURI() string {
	return string(config.LeetCodeCN) + "/"
}

func (c *cnClient) Inspect(typ string) (map[string]any, error) {
	query := `
query a {
  __type(name: "$type") {
    name 
    fields {
      name 
      args {
        name 
        description 
        defaultValue 
        type {
          name 
          kind 
          ofType {
            name 
            kind 
          }
        }
      }
      type {
        name 
        kind 
        ofType {
          name 
          kind 
        }
      }
    }
  }
}
`
	query = strings.ReplaceAll(query, "$type", typ)
	var resp map[string]any
	_, err := c.graphqlGet(
		graphqlRequest{query: query},
		&resp,
	)
	return resp, err
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

	body := struct {
		Login               string `url:"login"`
		Password            string `url:"password"`
		CsrfMiddlewareToken string `url:"csrfmiddlewaretoken"`
	}{username, password, csrfToken}
	req, err = c.http.New().Post(accountLoginPath).BodyForm(body).Request()
	if err != nil {
		return nil, err
	}
	resp, err = c.http.Do(req, nil, nil)
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
query globalData {
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
		graphqlRequest{query: query, authType: requireAuth}, &resp,
	)
	if err != nil {
		return nil, err
	}
	userStatus := resp.Data.UserStatus
	return &userStatus, nil
}

func (c *cnClient) getQuestionData(slug string, query string, authType authType) (*QuestionData, error) {
	var resp struct {
		Data struct {
			Question QuestionData `json:"question"`
		}
	}
	_, err := c.graphqlPost(
		graphqlRequest{
			query:         query,
			operationName: "questionData",
			variables:     map[string]any{"titleSlug": slug},
			authType:      authType,
		}, &resp,
	)
	if err != nil {
		return nil, err
	}
	q := resp.Data.Question
	if q.TitleSlug == "" {
		return nil, ErrQuestionNotFound
	}
	if q.IsPaidOnly && q.Content == "" {
		return nil, ErrPaidOnlyQuestion
	}
	return &q, nil
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
			exampleTestcaseList
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
			editorType
		}
	}`
	q, err := c.getQuestionData(slug, query, withAuth)
	if err != nil {
		return q, err
	}
	q.client = c
	return q, nil
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
			query:         query,
			operationName: "AllQuestionUrls",
		}, &resp,
	)
	if err != nil {
		return nil, err
	}
	url := resp.Get("data.allQuestionUrls.questionUrl").Str

	log.Debug("request", "url", url)
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
	var respErr UnexpectedStatusCode
	dec := progressDecoder{smartDecoder{LogResponse: false}, tracker}
	_, err = c.http.New().Get(url).ResponseDecoder(dec).Receive(&qs, &respErr)
	if err != nil {
		return nil, err
	}
	if respErr.IsError() {
		return nil, respErr
	}
	for i := range qs {
		qs[i].client = c
		qs[i].partial = 1
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
			query:         query,
			operationName: "questionOfToday",
			authType:      withoutAuth,
		}, &resp,
	)
	if err != nil {
		return nil, err
	}
	slug := resp.Get("data.todayRecord.0.question.titleSlug").Str
	return c.GetQuestionData(slug)
}

func (c *cnClient) GetQuestionOfDate(date time.Time) (*QuestionData, error) {
	query := `
	query dailyQuestionRecords($year: Int!, $month: Int!) {
	    dailyQuestionRecords(year: $year, month: $month) {
			date
			userStatus
			question {
	            titleSlug
	        }
	    }
	}`
	var resp gjson.Result
	_, err := c.graphqlPost(
		graphqlRequest{
			query: query,
			variables: map[string]any{
				"year":  date.Year(),
				"month": int(date.Month()),
			},
			authType: withAuth,
		},
		&resp,
	)
	if err != nil {
		return nil, err
	}
	dateStr := date.Format("2006-01-02")
	qs := resp.Get("data.dailyQuestionRecords").Array()
	for _, q := range qs {
		if q.Get("date").Str == dateStr {
			slug := q.Get("question.titleSlug").Str
			return c.GetQuestionData(slug)
		}
	}
	return nil, ErrQuestionNotFound
}

func (c *cnClient) getContest(contestSlug string) (*Contest, error) {
	path := fmt.Sprintf(contestInfoPath, contestSlug)
	var resp gjson.Result
	_, err := c.jsonGet(path, nil, withAuth, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Get("error").Exists() {
		return nil, errors.New(resp.Get("error").Str)
	}
	contestInfo := resp.Get("contest")
	contest := &Contest{
		Id:              int(contestInfo.Get("id").Int()),
		TitleSlug:       contestSlug,
		Title:           contestInfo.Get("title").Str,
		StartTime:       contestInfo.Get("start_time").Int(),
		OriginStartTime: contestInfo.Get("origin_start_time").Int(),
		Duration:        int(contestInfo.Get("duration").Int()),
		IsVirtual:       contestInfo.Get("is_virtual").Bool(),
		Description:     contestInfo.Get("description").Str,
		ContainsPremium: resp.Get("containsPremium").Bool(),
		Registered:      resp.Get("registered").Bool(),
		Questions:       make([]*QuestionData, 0, 4),
	}
	for _, q := range resp.Get("questions").Array() {
		question := &QuestionData{
			partial:         1,
			contest:         contest,
			TitleSlug:       q.Get("title_slug").Str,
			QuestionId:      q.Get("question_id").Str,
			Title:           q.Get("english_title").Str,
			TranslatedTitle: q.Get("title").Str,
		}
		contest.Questions = append(contest.Questions, question)
	}

	return contest, nil
}

func (c *cnClient) GetContest(contestSlug string) (*Contest, error) {
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

func (c *cnClient) GetContestQuestionData(contestSlug string, questionSlug string) (*QuestionData, error) {
	path := fmt.Sprintf(contestProblemsPath, contestSlug, questionSlug)
	var html []byte
	req, _ := c.http.New().Get(path).Request()
	_, err := c.send(req, requireAuth, &html)
	if err != nil {
		var e UnexpectedStatusCode
		if errors.As(err, &e) && e.Code == 302 {
			return nil, ErrPaidOnlyQuestion
		}
		return nil, err
	}
	if len(html) == 0 {
		return nil, errors.New("get contest question data: empty response")
	}
	q, err := parseContestHtml(html, questionSlug, config.LeetCodeCN)
	if err != nil {
		return nil, err
	}
	q.normalize()
	q.client = c
	return q, nil
}

func parseContestHtml(html []byte, questionSlug string, site config.LeetcodeSite) (*QuestionData, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, err
	}

	script := doc.Find("script#__NEXT_DATA__")
	if script.Length() == 0 {
		return nil, errors.New("get contest question data: empty script")
	}
	jsonText := script.Text()
	result := gjson.Get(jsonText, "props.pageProps.dehydratedState.queries.#.state.data.contestQuestion.question")
	if !result.Exists() {
		return nil, errors.New("contest question data not found in __NEXT_DATA__")
	}

	var questionRaw []byte
	for _, q := range result.Array() {
		if q.Exists() && len(q.Map()) > 0 {
			questionRaw = []byte(q.Raw)
			break
		}
	}
	if len(questionRaw) == 0 {
		return nil, errors.New("contest question data not found in __NEXT_DATA__")
	}

	var qd QuestionData
	err = json.Unmarshal(questionRaw, &qd)
	if err != nil {
		return nil, err
	}
	qd.TitleSlug = questionSlug
	return &qd, nil
}

// 每次 "运行代码" 会产生两个 submission, 一个是运行我们的代码，一个是运行标程。

// RunCode runs code on leetcode server. Questions no need to be fully loaded.
func (c *cnClient) RunCode(q *QuestionData, lang string, code string, dataInput string) (
	*InterpretSolutionResult,
	error,
) {
	path := ""
	if q.IsContest() {
		path = fmt.Sprintf(
			contestRunCodePath,
			q.contest.TitleSlug,
			q.TitleSlug,
		)
	} else {
		path = fmt.Sprintf(runCodePath, q.TitleSlug)
	}

	var resp InterpretSolutionResult
	_, err := c.jsonPost(
		path, map[string]any{
			"lang":        lang,
			"question_id": q.QuestionId,
			"typed_code":  code,
			"data_input":  dataInput,
		}, requireAuth, &resp,
	)
	if err != nil {
		return nil, err
	}
	return &resp, err
}

// SubmitCode submits code to leetcode server. Questions no need to be fully loaded.
func (c *cnClient) SubmitCode(q *QuestionData, lang string, code string) (string, error) {
	path := ""
	if q.IsContest() {
		path = fmt.Sprintf(
			contestSubmitCodePath,
			q.contest.TitleSlug,
			q.TitleSlug,
		)
	} else {
		path = fmt.Sprintf(submitCodePath, q.TitleSlug)
	}

	var resp gjson.Result
	_, err := c.jsonPost(
		path, map[string]any{
			"lang":         lang,
			"questionSlug": q.TitleSlug,
			"question_id":  q.QuestionId,
			"typed_code":   code,
		}, requireAuth, &resp,
	)
	return resp.Get("submission_id").String(), err
}

func (c *cnClient) CheckResult(submissionId string) (
	CheckResult,
	error,
) {
	path := fmt.Sprintf(checkResultPath, submissionId)
	var result gjson.Result
	_, err := c.jsonGet(path, nil, requireAuth, &result)
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

func (c *cnClient) GetUpcomingContests() ([]*Contest, error) {
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
		graphqlRequest{query: query, authType: withAuth}, &resp,
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

func (c *cnClient) RegisterContest(slug string) error {
	path := fmt.Sprintf(contestRegisterPath, slug)
	_, err := c.jsonPost(path, nil, requireAuth, nil)
	var e UnexpectedStatusCode
	if errors.As(err, &e) && e.Code == http.StatusFound {
		err = nil
	}
	return err
}

func (c *cnClient) UnregisterContest(slug string) error {
	path := fmt.Sprintf(contestRegisterPath, slug)
	req, _ := c.http.New().Delete(path).Request()
	_, err := c.send(req, requireAuth, nil)
	return err
}

type QuestionFilter struct {
	Difficulty     string   `json:"difficulty,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	Status         string   `json:"status,omitempty"`
	SearchKeywords string   `json:"searchKeywords,omitempty"`
}

func (c *cnClient) GetQuestionsByFilter(f QuestionFilter, limit int, skip int) (QuestionList, error) {
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
		}, &resp,
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

func (c *cnClient) GetQuestionTags() ([]QuestionTag, error) {
	var resp gjson.Result
	_, err := c.jsonGet(problemsApiTagsPath, nil, withAuth, &resp)
	if err != nil {
		return nil, err
	}
	var tags []QuestionTag
	for _, tag := range resp.Get("topics").Array() {
		tags = append(
			tags, QuestionTag{
				Slug:           tag.Get("slug").Str,
				Name:           tag.Get("name").Str,
				NameTranslated: tag.Get("translatedName").Str,
			},
		)
	}
	return tags, nil
}

func (c *cnClient) GetStreakCounter() (StreakCounter, error) {
	query := `
query getStreakCounter {
  problemsetStreakCounter {
    today
    streakCount
    daysSkipped
    todayCompleted
  }
}`
	var resp gjson.Result
	_, err := c.graphqlPost(
		graphqlRequest{path: graphQLNoj, query: query, authType: requireAuth}, &resp,
	)
	if err != nil {
		return StreakCounter{}, err
	}
	var counter StreakCounter
	err = json.Unmarshal(utils.StringToBytes(resp.Get("data.problemsetStreakCounter").Raw), &counter)
	return counter, err
}
