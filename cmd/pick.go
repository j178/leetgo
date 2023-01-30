package cmd

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/j178/leetgo/editor"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
)

func askFilter(c leetcode.Client) (filter leetcode.QuestionFilter, err error) {
	tags, err := c.GetQuestionTags()
	if err != nil {
		return
	}
	var tagNames []string
	for _, t := range tags {
		tagNames = append(tagNames, t.Name)
	}

	qs := []*survey.Question{
		{
			Name: "Difficulty",
			Prompt: &survey.Select{
				Message: "Select a difficulty level",
				Options: []string{"All", "Easy", "Medium", "Hard"},
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				opt := ans.(survey.OptionAnswer)
				if opt.Index == 0 {
					return survey.OptionAnswer{Value: ""}
				}
				opt.Value = strings.ToUpper(opt.Value)
				return opt
			},
		},
		{
			Name: "Status",
			Prompt: &survey.Select{
				Message: "Select question status",
				Options: []string{"All", "Not Started", "Tried", "Ac"},
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				opt := ans.(survey.OptionAnswer)
				if opt.Index == 0 {
					return survey.OptionAnswer{Value: ""}
				}
				opt.Value = strings.ReplaceAll(strings.ToUpper(opt.Value), " ", "_")
				return opt
			},
		},
		{
			Name: "Tags",
			Prompt: &survey.MultiSelect{
				Message: "Select tags",
				Options: tagNames,
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				opt := ans.([]survey.OptionAnswer)
				if len(opt) == 0 {
					return opt
				}
				if len(opt) == len(tagNames) {
					return []survey.OptionAnswer{}
				}
				return ans
			},
		},
	}

	err = survey.Ask(qs, &filter, survey.WithRemoveSelectAll())
	if err != nil {
		return
	}

	return filter, nil
}

var pickCmd = &cobra.Command{
	Use:   "pick [qid]",
	Short: "Generate a new question",
	Example: `leetgo pick  # show a list of questions to pick
leetgo pick today
leetgo pick 549
leetgo pick two-sum`,
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
				return fmt.Errorf("`leetgo pick` cannot handle multiple contest questions, use `leetgo contest` instead.")
			}
			q = qs[0]
		} else {
			filter, err := askFilter(c)
			if err != nil {
				return err
			}
			m := newTuiModel(filter, c)
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
