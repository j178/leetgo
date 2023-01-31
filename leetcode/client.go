package leetcode

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dghubble/sling"
	"github.com/goccy/go-json"
	"github.com/hashicorp/go-hclog"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/tidwall/gjson"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/utils"
)

var (
	ErrPaidOnlyQuestion  = errors.New("this is paid only question, you need to subscribe to LeetCode Premium")
	ErrTooManyRequests   = errors.New("you have submitted too frequently, please submit again later")
	ErrQuestionNotFound  = errors.New("no such question")
	ErrContestNotStarted = errors.New("contest has not started")
	ErrUserNotSignedIn   = errors.New("user not signed in")
)

type unexpectedStatusCode struct {
	Code int
	Resp *http.Response
	Body []byte
}

func (e unexpectedStatusCode) Error() string {
	body := "<empty>"
	if len(e.Body) > 0 {
		body = string(e.Body)
	}
	return fmt.Sprintf("unexpected status code: %d, body: %s", e.Code, body)
}

// nolint: unused
type errorResponse struct {
	ErrorMsg *string `json:"error"`
}

// nolint: unused
func (e errorResponse) Ok() bool {
	return e.ErrorMsg == nil
}

// nolint: unused
func (e errorResponse) Error() string {
	if e.ErrorMsg != nil {
		return *e.ErrorMsg
	}
	return ""
}

type Client interface {
	BaseURI() string
	Inspect(typ string) (map[string]any, error)
	Login(username, password string) (*http.Response, error)
	GetUserStatus() (*UserStatus, error)
	GetQuestionData(slug string) (*QuestionData, error)
	GetAllQuestions() ([]*QuestionData, error)
	GetTodayQuestion() (*QuestionData, error)
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
	httpClient.Client(
		&http.Client{
			CheckRedirect: nonFollowRedirect,
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
	query         string
	operationName string
	variables     map[string]any
}

const (
	graphQLPath           = "/graphql"
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
	problemSetAllPath     = "/problemset/all/"
)

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

	// default error detection
	resp, err := c.http.Do(req, result, failure)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return resp, ErrTooManyRequests
	}

	if !(200 <= resp.StatusCode && resp.StatusCode <= 299) {
		body, _ := io.ReadAll(resp.Body)
		return resp, unexpectedStatusCode{Code: resp.StatusCode, Resp: resp, Body: body}
	}

	return nil, err
}

//nolint:unused
func (c *cnClient) graphqlGet(req graphqlRequest, result any, failure any) (*http.Response, error) {
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
	r, err := c.http.New().Get(graphQLPath).QueryStruct(p).Request()
	if err != nil {
		return nil, err
	}
	return c.send(r, result, failure)
}

func (c *cnClient) graphqlPost(req graphqlRequest, result any, failure any) (*http.Response, error) {
	v := req.variables
	if v == nil {
		v = make(map[string]any)
	}
	body := map[string]any{
		"query":         req.query,
		"operationName": req.operationName,
		"variables":     v,
	}
	r, err := c.http.New().Post(graphQLPath).BodyJSON(body).Request()
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

// same
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
		nil,
	)
	return resp, err
}

// only cn
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

// same
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
		graphqlRequest{query: query}, &resp, nil,
	)
	if err != nil {
		return nil, err
	}
	userStatus := resp.Data.UserStatus
	return &userStatus, nil
}

func (c *cnClient) getQuestionData(slug string, query string) (*QuestionData, error) {
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
		}, &resp, nil,
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

// minor difference
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
		}
	}`
	q, err := c.getQuestionData(slug, query)
	if err != nil {
		return q, err
	}
	q.client = c
	return q, nil
}

// different
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

// different
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
		}, &resp, nil,
	)
	if err != nil {
		return nil, err
	}
	slug := resp.Get("data.todayRecord.0.question.titleSlug").Str
	return c.GetQuestionData(slug)
}

// same
func (c *cnClient) getContest(contestSlug string) (*Contest, error) {
	path := fmt.Sprintf(contestInfoPath, contestSlug)
	var resp gjson.Result
	_, err := c.jsonGet(path, nil, &resp, nil)
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

// same
func (c *cnClient) getContestQuestionData(contestSlug string, questionSlug string) (*QuestionData, error) {
	path := fmt.Sprintf(contestProblemsPath, contestSlug, questionSlug)
	var html []byte
	req, _ := c.http.New().Get(path).Request()
	_, err := c.send(req, &html, nil)
	if err != nil {
		return nil, err
	}
	if len(html) == 0 {
		return nil, errors.New("get contest question data: empty response")
	}

	return c.parseContestHtml(html, questionSlug)
}

func (c *cnClient) GetContestQuestionData(contestSlug string, questionSlug string) (*QuestionData, error) {
	q, err := c.getContestQuestionData(contestSlug, questionSlug)
	if err != nil {
		return nil, err
	}
	q.client = c
	return q, nil
}

func (c *cnClient) parseContestHtml(html []byte, questionSlug string) (*QuestionData, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, err
	}
	difficulty := strings.TrimSpace(doc.Find("span.pull-right.label.round").Text())
	frontendId := strings.TrimSuffix(doc.Find("div.question-title h3").Get(0).FirstChild.Data, ". ")
	translatedContent, err := doc.Find("div.question-content.default-content").Html()
	if err != nil {
		return nil, err
	}
	content, err := doc.Find("div.question-content.source-content").Html()
	if err != nil {
		return nil, err
	}

	var (
		questionId         string
		scriptText         string
		codeDefinitionText string
		metaDataText       string
		title              string
		sourceTitle        string
		exampleTestcases   string
		sampleTestcase     string
		categoryTitle      string
	)
	for _, node := range doc.Find("script").Nodes {
		if node.FirstChild != nil && strings.Contains(node.FirstChild.Data, "var pageData") {
			scriptText = node.FirstChild.Data
			break
		}
	}
	if scriptText == "" {
		return nil, errors.New("question data not found")
	}
	scriptLines := strings.Split(scriptText, "\n")
	for _, line := range scriptLines {
		switch {
		case strings.HasPrefix(line, `    questionId: '`):
			questionId = line[len(`    questionId: '`) : len(line)-2]
		case strings.HasPrefix(line, `    questionTitle: '`):
			title = line[len(`    questionTitle: '`) : len(line)-2]
		case strings.HasPrefix(line, `    questionSourceTitle: '`):
			sourceTitle = line[len(`    questionSourceTitle: '`) : len(line)-2]
		case strings.HasPrefix(line, `    questionExampleTestcases: '`):
			exampleTestcases = line[len(`    questionExampleTestcases: '`) : len(line)-2]
			exampleTestcases = utils.DecodeRawUnicodeEscape(exampleTestcases)
		case strings.HasPrefix(line, `    sampleTestCase: '`):
			sampleTestcase = line[len(`    sampleTestCase: '`) : len(line)-2]
			sampleTestcase = utils.DecodeRawUnicodeEscape(sampleTestcase)
		case strings.HasPrefix(line, `    codeDefinition: `):
			codeDefinitionText = line[len(`    codeDefinition: `):len(line)-len(",],")] + "]"
			codeDefinitionText = strings.ReplaceAll(codeDefinitionText, "'", `"`)
		case strings.HasPrefix(line, `    metaData: `):
			metaDataText = line[len(`    metaData: JSON.parse('`) : len(line)-len(`' || '{}'),`)]
			metaDataText = utils.DecodeRawUnicodeEscape(metaDataText)
		case strings.HasPrefix(line, `    categoryTitle: '`):
			categoryTitle = line[len(`    categoryTitle: '`) : len(line)-2]
		}
	}

	q := &QuestionData{
		QuestionId:         questionId,
		QuestionFrontendId: frontendId,
		TitleSlug:          questionSlug,
		Difficulty:         difficulty,
		Content:            content,
		TranslatedContent:  translatedContent,
		Title:              sourceTitle,
		TranslatedTitle:    title,
		ExampleTestcases:   exampleTestcases,
		SampleTestCase:     sampleTestcase,
		CategoryTitle:      CategoryTitle(categoryTitle),
	}
	err = json.Unmarshal([]byte(metaDataText), &q.MetaData)
	if err != nil {
		return nil, err
	}
	var codeDefs []map[string]string
	err = json.Unmarshal([]byte(codeDefinitionText), &codeDefs)
	if err != nil {
		return nil, err
	}
	for _, codeDef := range codeDefs {
		q.CodeSnippets = append(
			q.CodeSnippets, CodeSnippet{
				LangSlug: codeDef["value"],
				Lang:     codeDef["text"],
				Code:     codeDef["defaultCode"],
			},
		)
	}
	return q, nil
}

// 每次 "运行代码" 会产生两个 submission, 一个是运行我们的代码，一个是运行标程。

// same
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
		}, &resp, nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp, err
}

// same
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
		}, &resp, nil,
	)
	return resp.Get("submission_id").String(), err
}

// same
func (c *cnClient) CheckResult(submissionId string) (
	CheckResult,
	error,
) {
	path := fmt.Sprintf(checkResultPath, submissionId)
	var result gjson.Result
	_, err := c.jsonGet(path, nil, &result, nil)
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

// different
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

// same
func (c *cnClient) RegisterContest(slug string) error {
	path := fmt.Sprintf(contestRegisterPath, slug)
	_, err := c.jsonPost(path, nil, nil, nil)
	if e, ok := err.(unexpectedStatusCode); ok && e.Code == http.StatusFound {
		err = nil
	}
	return err
}

// same
func (c *cnClient) UnregisterContest(slug string) error {
	path := fmt.Sprintf(contestRegisterPath, slug)
	req, _ := c.http.New().Delete(path).Request()
	_, err := c.send(req, nil, nil)
	return err
}

type QuestionFilter struct {
	Difficulty     string   `json:"difficulty,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	Status         string   `json:"status,omitempty"`
	SearchKeywords string   `json:"searchKeywords,omitempty"`
}

// different
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

// different
func (c *cnClient) GetQuestionTags() ([]QuestionTag, error) {
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
