package leetcode

import (
	"net/http"

	"github.com/dghubble/sling"
	"github.com/j178/leetgo/config"
)

type usClient struct {
	opt  Options
	http *sling.Sling
}

func (c *usClient) BaseURI() string {
	return string(config.LeetCodeUS) + "/"
}

func (c *usClient) Login(username, password string) (*http.Response, error) {
	// TODO implement me
	panic("implement me")
}

func (c *usClient) GetUserStatus() (*UserStatus, error) {
	// TODO implement me
	panic("implement me")
}

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
