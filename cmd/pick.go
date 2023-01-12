package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/j178/leetgo/editor"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

func askFilter() (leetcode.QuestionFilter, error) {
	var filter leetcode.QuestionFilter
	qs := []*survey.Question{
		{
			Name: "Difficulty",
			Prompt: &survey.Select{
				Message: "Select a difficulty level",
				Options: []string{"Easy", "Medium", "Hard"},
			},
		},
		{
			Name: "Status",
			Prompt: &survey.Select{
				Message: "Select question status",
				Options: []string{"Not Started", "Tried", "Ac"},
			},
		},
		{
			Name: "Tags",
			Prompt: &survey.MultiSelect{
				Message: "Select tags",
				Options: []string{},
			},
		},
		{
			Name:   "SearchKeywords",
			Prompt: &survey.Input{Message: "Keyword to search"},
		},
	}

	err := survey.Ask(qs, &filter)
	if err != nil {
		return filter, err
	}

	return filter, nil
}

type questionLoader struct {
	c         leetcode.Client
	filter    leetcode.QuestionFilter
	questions []*leetcode.QuestionData
}

var pickCmd = &cobra.Command{
	Use:   "pick [qid]",
	Short: "Generate a new question",
	Example: `leetgo pick  # show a list of questions to pick
leetgo pick today
leetgo pick 549`,
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"p"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		var q *leetcode.QuestionData

		if len(args) > 0 {
			qid := args[0]
			qs, err := leetcode.ParseQID(qid, c)
			if err != nil {
				return err
			}
			if len(qs) > 1 {
				return fmt.Errorf("multiple questions found")
			}
			q = qs[0]
		} else {
			cache := leetcode.GetCache(c)
			m := newTuiModel(cache)
			p := tea.NewProgram(m)
			if _, err := p.Run(); err != nil {
				return err
			}
			if m.Selected() == nil {
				return nil
			}
			q = m.Selected()
		}

		result, err := lang.Generate(q)
		if err != nil {
			return err
		}
		err = editor.Open(result.Files)
		return err
	},
}
