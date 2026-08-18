package leetcode

import (
	"testing"

	"github.com/j178/leetgo/config"
)

func TestQuestionFromSavedStateWithoutMetadata(t *testing.T) {
	saved := config.LastQuestion{
		Slug:       "two-sum",
		FrontendID: "1",
	}
	state := config.State{
		Questions: map[string]config.LastQuestion{
			"two-sum": saved,
		},
	}

	q, ok := questionFromSavedQuestions(nil, state.Questions, "1")
	if !ok {
		t.Fatal("questionFromSavedQuestions() did not find the saved question")
	}
	if q.TitleSlug != "two-sum" || q.QuestionFrontendId != "1" {
		t.Fatalf("unexpected question: %+v", q)
	}
}

func TestContestFromSavedStateByNumber(t *testing.T) {
	state := config.State{
		Questions: map[string]config.LastQuestion{
			"two-sum": {
				Slug:        "two-sum",
				FrontendID:  "1",
				ContestSlug: "weekly-contest-330",
				MetaData:    []byte(`{}`),
			},
		},
		Contests: map[string][]string{
			"weekly-contest-330": {"two-sum"},
		},
	}

	contest, questions, ok := contestFromSavedState(nil, state, "weekly-contest-330", 1, true)
	if !ok {
		t.Fatal("contestFromSavedState() did not find the saved contest")
	}
	if contest.TitleSlug != "weekly-contest-330" || len(questions) != 1 {
		t.Fatalf("unexpected contest result: contest=%+v questions=%d", contest, len(questions))
	}
	if questions[0].TitleSlug != "two-sum" || questions[0].Contest() != contest {
		t.Fatalf("question was not attached to the saved contest: %+v", questions[0])
	}
}

func TestContestFromSavedStateAllQuestions(t *testing.T) {
	state := config.State{
		Questions: map[string]config.LastQuestion{
			"two-sum": {Slug: "two-sum", MetaData: []byte(`{}`)},
			"add-two": {Slug: "add-two", MetaData: []byte(`{}`)},
		},
		Contests: map[string][]string{
			"weekly-contest-330": {"two-sum", "add-two"},
		},
	}

	_, questions, ok := contestFromSavedState(nil, state, "weekly-contest-330", -1, true)
	if !ok || len(questions) != 2 {
		t.Fatalf("contestFromSavedState() = ok:%v questions:%d, want two questions", ok, len(questions))
	}
}
