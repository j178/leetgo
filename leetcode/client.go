package leetcode

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "reflect"

    "github.com/dghubble/sling"
    "github.com/tidwall/gjson"
)

const (
    originCN = "https://leetcode.cn"
    originEN = "https://leetcode.com"
)

type Client interface {
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

var debugResponse = false

func (d decoder) Decode(resp *http.Response, v interface{}) error {
    data, _ := io.ReadAll(resp.Body)
    if debugResponse {
        fmt.Println(string(data))
    }
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
    baseUri string
    opts    Options
    http    *sling.Sling
}

func NewClient(options ...Option) Client {
    c := &cnClient{
        baseUri: originCN,
        http:    sling.New(),
    }
    for _, f := range options {
        f(&c.opts)
    }

    c.http.Base(c.baseUri)
    c.http.Add(
        "User-Agent",
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36",
    )
    c.http.Add("Accept-Encoding", "gzip, deflate, br")
    c.http.Add("Referer", c.baseUri)
    c.http.Add("Origin", c.baseUri[:len(c.baseUri)-1])
    c.http.ResponseDecoder(decoder{})
    return c
}

type graphQLBody struct {
    Query         string    `url:"query" json:"query"`
    OperationName string    `url:"operationName" json:"operationName"`
    Variables     Variables `url:"variables" json:"variables"`
}

func (c *cnClient) graphqlGet(query string, operation string, variables Variables) *sling.Sling {
    r := c.http.New().Get("/graphql/").QueryStruct(
        &graphQLBody{
            Query:         query,
            OperationName: operation,
            Variables:     variables,
        },
    )
    return r
}

func (c *cnClient) graphqlPost(query string, operation string, variables Variables) *sling.Sling {
    r := c.http.New().Post("/graphql/").BodyJSON(
        &graphQLBody{
            Query:         query,
            OperationName: operation,
            Variables:     variables,
        },
    )
    return r
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
    _, err := c.graphqlPost(query, "questionData", Variables{"titleSlug": slug}).ReceiveSuccess(&resp)
    if err != nil {
        return QuestionData{}, err
    }
    if resp.Data.Question.TitleSlug == "" {
        return QuestionData{}, errors.New("question not found")
    }
    return resp.Data.Question, nil
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
    _, err := c.graphqlPost(query, "AllQuestionUrls", nil).ReceiveSuccess(&resp)
    if err != nil {
        return nil, err
    }
    url := resp.Get("data.allQuestionUrls.questionUrl").Str

    var qs []QuestionData
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
    _, err := c.graphqlPost(query, "questionOfToday", nil).ReceiveSuccess(&resp)
    if err != nil {
        return QuestionData{}, err
    }
    slug := resp.Get("data.todayRecord.0.question.titleSlug").Str
    return c.GetQuestionData(slug)
}
