package leetcode

type Contest struct {
	TitleSlug string `json:"titleSlug"`
	Title     string `json:"title"`
	StartTime string `json:"startTime"`
	Questions []*QuestionData
}

func ContestBySlug(slug string, c Client) *Contest {
	return nil
}
