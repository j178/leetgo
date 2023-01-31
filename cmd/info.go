package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/j178/leetgo/leetcode"
)

var infoCmd = &cobra.Command{
	Use:     "info qid...",
	Short:   "Show question info",
	Example: "leetgo info 145\nleetgo info two-sum",
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"i"},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient()
		var questions []*leetcode.QuestionData

		for _, qid := range args {
			qs, err := leetcode.ParseQID(qid, c)
			if err != nil {
				hclog.L().Error("failed to get question", "qid", qid, "err", err)
				continue
			}
			questions = append(questions, qs...)
		}
		if len(questions) == 0 {
			return errors.New("no questions found")
		}

		w := table.NewWriter()
		w.SetOutputMirror(os.Stdout)
		w.SetStyle(table.StyleColoredDark)
		w.SetColumnConfigs(
			[]table.ColumnConfig{
				{
					Number:   2,
					WidthMax: 50,
				},
			},
		)
		for _, q := range questions {
			_ = q.Fulfill()
			w.AppendRow(
				table.Row{fmt.Sprintf("%s. %s", q.QuestionFrontendId, q.GetTitle())},
				table.RowConfig{AutoMerge: true, AutoMergeAlign: text.AlignLeft},
			)
			w.AppendRow(table.Row{"Slug", q.TitleSlug})
			w.AppendRow(table.Row{"Difficulty", q.Difficulty})
			w.AppendRow(table.Row{"URL", q.Url()})
			w.AppendRow(table.Row{"Tags", strings.Join(q.TagSlugs(), ", ")})
			w.AppendRow(table.Row{"Paid Only", q.IsPaidOnly})
			w.AppendRow(
				table.Row{
					"AC Rate",
					fmt.Sprintf("%s/%s %s", q.Stats.TotalAccepted, q.Stats.TotalSubmission, q.Stats.ACRate),
				},
			)
			w.AppendRow(table.Row{"Content", q.GetFormattedContent()})
			for _, h := range q.Hints {
				w.AppendRow(table.Row{"Hint", h})
			}
			w.AppendSeparator()
		}
		w.Render()
		return nil
	},
}
