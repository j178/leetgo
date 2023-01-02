package leetcode

import (
	"errors"
	"net/http"

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

// GetUserStatus can be reused from cnClient

func (c *usClient) GetQuestionData(slug string) (*QuestionData, error) {
	// TODO implement me
	panic("implement me")
}

func (c *usClient) GetAllQuestions() ([]*QuestionData, error) {
	// TODO implement me
	panic("implement me")
}

func (c *usClient) GetTodayQuestion() (*QuestionData, error) {
	// TODO implement me
	panic("implement me")
}

func (c *usClient) InterpretSolution(
	q *QuestionData,
	lang string,
	code string,
	dataInput string,
) (*InterpretSolutionResult, error) {
	// TODO implement me
	panic("implement me")
}

func (c *usClient) CheckSubmissionResult(submissionId string) (*SubmissionCheckResult, error) {
	// TODO implement me
	panic("implement me")
}

func (c *usClient) Submit(q *QuestionData, lang string, code string) (string, error) {
	// TODO implement me
	panic("implement me")
}
