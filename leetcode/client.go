package leetcode

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/dghubble/sling"
	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
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

type decoder struct {
	path string
}

func (d decoder) Decode(resp *http.Response, v interface{}) error {
	data, _ := io.ReadAll(resp.Body)
	hclog.L().Trace("Leetcode response", "data", string(data), "url", resp.Request.URL.String())

	ty := reflect.TypeOf(v)
	ele := reflect.ValueOf(v).Elem()
	switch ty.Elem() {
	case reflect.TypeOf(gjson.Result{}):
		if d.path == "" {
			ele.Set(reflect.ValueOf(gjson.ParseBytes(data)))
		} else {
			ele.Set(reflect.ValueOf(gjson.GetBytes(data, d.path)))
		}
	case reflect.TypeOf([]byte{}):
		ele.SetBytes(data)
	default:
		return json.Unmarshal(data, v)
	}
	return nil
}

type ErrorResp struct {
	Errors string `json:"errors"`
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
	httpClient.ResponseDecoder(decoder{})

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
	hclog.L().Trace("Requesting", "method", "GET", "url", r.URL.String())
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
	hclog.L().Trace("Requesting", "method", "POST", "url", r.URL.String())
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

	var qs []QuestionData
	hclog.L().Trace("Requesting", "url", url)
	_, err = c.http.New().Get(url).ReceiveSuccess(&qs)
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
