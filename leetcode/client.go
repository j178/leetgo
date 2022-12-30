package leetcode

import (
	"errors"
	"fmt"

	"github.com/dghubble/sling"
	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/tidwall/gjson"
)

type Client interface {
	BaseURI() string
	GetQuestionData(slug string) (QuestionData, error)
	GetAllQuestions() ([]QuestionData, error)
	GetTodayQuestion() (QuestionData, error)
}

type Option func(opts *Options)

type Options struct {
	cred CredentialProvider
}

func WithCredential(cred CredentialProvider) Option {
	return func(opts *Options) {
		opts.cred = cred
	}
}

type Variables map[string]string

type cnClient struct {
	opts Options
	http *sling.Sling
}

func NewClient(options ...Option) Client {
	var opts Options
	for _, f := range options {
		f(&opts)
	}
	httpClient := sling.New()
	httpClient.Add(
		"User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36",
	)
	httpClient.Add("Accept-Encoding", "gzip, deflate, br")
	httpClient.ResponseDecoder(smartDecoder{LogResponseData: true})

	cfg := config.Get()
	if cfg.LeetCode.Site == config.LeetCodeCN {
		c := &cnClient{
			opts: opts,
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

type graphQLBody struct {
	Query         string    `url:"query" json:"query"`
	OperationName string    `url:"operationName" json:"operationName"`
	Variables     Variables `url:"variables" json:"variables"`
}

//nolint:unused
func (c *cnClient) graphqlGet(query string, operation string, variables Variables, result any) error {
	r, err := c.http.New().Get("/graphql/").QueryStruct(
		&graphQLBody{
			Query:         query,
			OperationName: operation,
			Variables:     variables,
		},
	).Request()
	if err != nil {
		return err
	}
	hclog.L().Trace("request", "method", "GET", "url", r.URL.String())
	_, err = c.http.Do(r, result, nil)
	return err
}

func (c *cnClient) graphqlPost(query string, operation string, variables Variables, result any) error {
	r, err := c.http.New().Post("/graphql/").BodyJSON(
		&graphQLBody{
			Query:         query,
			OperationName: operation,
			Variables:     variables,
		},
	).Request()
	if err != nil {
		return err
	}
	hclog.L().Trace("request", "method", "POST", "url", r.URL.String())
	_, err = c.http.Do(r, result, nil)
	return err
}

func (c *cnClient) BaseURI() string {
	return string(config.LeetCodeCN) + "/"
}

func (c *cnClient) GetQuestionData(slug string) (QuestionData, error) {
	query := `
	query questionData($titleSlug: String!) {
		question(titleSlug: $titleSlug) {
			questionId
			questionFrontendId
			title
			titleSlug
			content
			isPaidOnly
			translatedTitle
			translatedContent
			difficulty
			stats
			hints
			similarQuestions
			sampleTestCase
			exampleTestcases
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
	err := c.graphqlPost(query, "questionData", Variables{"titleSlug": slug}, &resp)
	if err != nil {
		return QuestionData{}, err
	}
	if resp.Data.Question.TitleSlug == "" {
		return QuestionData{}, errors.New("question not found")
	}
	q := resp.Data.Question
	q.client = c
	return q, nil
}

func (c *cnClient) GetAllQuestions() ([]QuestionData, error) {
	query := `
	query AllQuestionUrls {
		allQuestionUrls {
			questionUrl
		}
	}
	`
	var resp gjson.Result
	err := c.graphqlPost(query, "AllQuestionUrls", nil, &resp)
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

	var qs []QuestionData
	dec := progressDecoder{smartDecoder{LogResponseData: false}, tracker}
	_, err = c.http.New().Get(url).ResponseDecoder(dec).ReceiveSuccess(&qs)
	if err != nil {
		return nil, err
	}
	return qs, err
}

func (c *cnClient) GetTodayQuestion() (QuestionData, error) {
	query := `
    query questionOfToday {
        todayRecord {
            question {
                titleSlug
            }
        }
    }`
	var resp gjson.Result
	err := c.graphqlPost(query, "questionOfToday", nil, &resp)
	if err != nil {
		return QuestionData{}, err
	}
	slug := resp.Get("data.todayRecord.0.question.titleSlug").Str
	return c.GetQuestionData(slug)
}
