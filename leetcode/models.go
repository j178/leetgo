package leetcode

type Question struct {
	TitleSlug          string `json:"titleSlug"`
	Title              string `json:"title"`
	QuestionId         string `json:"questionId"`
	QuestionFrontendId string `json:"questionFrontendId"`
}

type ErrorResp struct {
	Errors string `json:"errors"`
}

type Variables map[string]string
