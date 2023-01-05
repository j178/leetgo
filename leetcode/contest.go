package leetcode

type Contest struct {
	TitleSlug string `json:"titleSlug"`
	Title     string `json:"title"`
	StartTime string `json:"startTime"`
	Questions []*QuestionData
}

func (ct *Contest) GetQuestionByNumber(num int, c Client) (*QuestionData, error) {
	return nil, nil
}

func (ct *Contest) GetAllQuestions(c Client) ([]*QuestionData, error) {
	return nil, nil
}
