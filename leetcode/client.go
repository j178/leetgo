package leetcode

import (
	"fmt"
	"io"
	"net/http"

	"github.com/dghubble/sling"
)

const (
	originCN = "https://leetcode.cn"
	originEN = "https://leetcode.com"
)

type Option func(*Client)

func WithEn() Option {
	return func(c *Client) {
		c.cn = false
		c.baseUri = originEN + "/"
	}
}

func WithCredential(cred CredentialProvider) Option {
	return func(c *Client) {
		c.cred = cred
	}
}

type Client struct {
	cn      bool
	baseUri string
	cred    CredentialProvider
	http    *sling.Sling
}

func NewClient(options ...Option) *Client {
	c := &Client{
		cn:      true,
		baseUri: originCN,
		http:    sling.New(),
	}
	for _, f := range options {
		f(c)
	}

	c.http.Base(c.baseUri)
	c.http.Add(
		"User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36",
	)
	c.http.Add("Accept-Encoding", "gzip, deflate, br")
	c.http.Add("Referer", c.baseUri)
	c.http.Add("Origin", c.baseUri[:len(c.baseUri)-1])

	return c
}

type graphQLBody struct {
	Query         string            `url:"query" json:"query"`
	OperationName string            `url:"operationName" json:"operationName"`
	Variables     map[string]string `url:"variables" json:"variables"`
}

func (c *Client) graphqlGet(query string, operation string, variables Variables) *sling.Sling {
	r := c.http.New().Get("/graphql/").QueryStruct(
		&graphQLBody{
			Query:         query,
			OperationName: operation,
			Variables:     variables,
		},
	)
	return r
}

func (c *Client) graphqlPost(query string, operation string, variables Variables) *sling.Sling {
	r := c.http.New().Post("/graphql/").BodyJSON(
		&graphQLBody{
			Query:         query,
			OperationName: operation,
			Variables:     variables,
		},
	)
	return r
}

func (c *Client) GetQuestionData(slug string) (Question, error) {
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
	var q struct {
		Data struct {
			Question `json:"question"`
		} `json:"data"`
	}
	_, err := c.graphqlPost(query, "questionData", Variables{"titleSlug": slug}).ReceiveSuccess(&q)
	if err != nil {
		return Question{}, err
	}
	return q.Data.Question, nil
}

type debugRespDecoder struct{}

func (debugRespDecoder) Decode(resp *http.Response, v interface{}) error {
	data, _ := io.ReadAll(resp.Body)
	fmt.Println(string(data))
	return nil
}

func (c *Client) GetAllQuestions() ([]map[string]any, error) {
	query := `
	query AllQuestionUrls {
		allQuestionUrls {
			questionUrl
		}
	}
	`
	var resp struct {
		Data struct {
			AllQuestionUrls struct {
				QuestionUrl string `json:"questionUrl"`
			} `json:"allQuestionUrls"`
		} `json:"data"`
	}
	_, err := c.graphqlPost(query, "AllQuestionUrls", nil).ReceiveSuccess(&resp)
	if err != nil {
		return nil, err
	}
	url := resp.Data.AllQuestionUrls.QuestionUrl

	var all []map[string]any
	_, err = sling.New().Get(url).ReceiveSuccess(&all)
	if err != nil {
		return nil, err
	}
	return all, err
}
