package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"encoding/json"

	"github.com/charmbracelet/log"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/j178/leetgo/leetcode"
)

type outputFormatEnum string

const (
    jsonOutput outputFormatEnum = "json"
)

func (e *outputFormatEnum) String() string {
    return string(*e)
}

func (e *outputFormatEnum) Set(v string) error {
    switch v {
    case "json":
        *e = outputFormatEnum(v)
        return nil
    default:
        return errors.New(`must be one of the support formats: "json"`)
    }
}

func (e *outputFormatEnum) Type() string {
    return "outputFormatEnum"
}

var flagFull bool
var flagFormat outputFormatEnum

func init() {
	infoCmd.Flags().BoolVarP(&flagFull, "full", "f", false, "show full question info")
	infoCmd.Flags().Var(&flagFormat, "format", "preseent question info in raw json string")
}

var infoCmd = &cobra.Command{
	Use:       "info qid...",
	Short:     "Show question info",
	Example:   "leetgo info 145\nleetgo info two-sum",
	Args:      cobra.MinimumNArgs(1),
	Aliases:   []string{"i"},
	ValidArgs: []string{"today", "last"},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient(leetcode.ReadCredentials())
		var questions []*leetcode.QuestionData

		for _, qid := range args {
			qs, err := leetcode.ParseQID(qid, c)
			if err != nil {
				log.Error("failed to get question", "qid", qid, "err", err)
				continue
			}
			questions = append(questions, qs...)
		}
		if len(questions) == 0 {
			return errors.New("no questions found")
		}

		if len(flagFormat) > 0 {
			res, err := json.Marshal(questions)
			if err != nil {
				return fmt.Errorf("failed to convert questions to raw string", "err", err)
			}
			fmt.Println(string(res))
			return nil
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
			if flagFull {
				w.AppendRow(table.Row{"Content", q.GetFormattedContent()})
			}
			for _, h := range q.Hints {
				w.AppendRow(table.Row{"Hint", h})
			}
			w.AppendSeparator()
		}
		w.Render()
		return nil
	},
}
