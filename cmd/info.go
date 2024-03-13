package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/j178/leetgo/leetcode"
)

type outputFormat string

func (e *outputFormat) String() string {
	return string(*e)
}

func (e *outputFormat) Set(v string) error {
	switch v {
	case "json":
		*e = outputFormat(v)
		return nil
	default:
		return errors.New(`must be one of the supported formats: "json"`)
	}
}

func (e *outputFormat) Type() string {
	return "outputFormat"
}

var (
	flagFull   bool
	flagFormat outputFormat = "default"
)

func init() {
	infoCmd.Flags().BoolVar(&flagFull, "full", false, "show full question info")
	infoCmd.Flags().Var(&flagFormat, "format", "show question info in specific format (json)")
}

// A simplified version of the leetcode.QuestionData struct
type question struct {
	FrontendId         string   `json:"frontend_id"`
	Title              string   `json:"title"`
	Slug               string   `json:"slug"`
	Difficulty         string   `json:"difficulty"`
	Url                string   `json:"url"`
	Tags               []string `json:"tags"`
	IsPaidOnly         bool     `json:"is_paid_only"`
	TotalAccepted      string   `json:"total_accepted"`
	TotalAcceptedRaw   int      `json:"total_accepted_raw"`
	TotalSubmission    string   `json:"total_submission"`
	TotalSubmissionRaw int      `json:"total_submission_raw"`
	ACRate             string   `json:"ac_rate"`
	Content            string   `json:"content"`
	Hints              []string `json:"hints"`
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

		var questions []question
		for _, qid := range args {
			qs, err := leetcode.ParseQID(qid, c)
			if err != nil {
				log.Error("failed to get question", "qid", qid, "err", err)
				continue
			}
			for _, q := range qs {
				_ = q.Fulfill()
				content := ""
				if flagFull {
					content = q.GetFormattedContent()
				}
				questions = append(
					questions, question{
						FrontendId:         q.QuestionFrontendId,
						Title:              q.GetTitle(),
						Slug:               q.TitleSlug,
						Difficulty:         q.Difficulty,
						Url:                q.Url(),
						Tags:               q.TagSlugs(),
						IsPaidOnly:         q.IsPaidOnly,
						TotalAccepted:      q.Stats.TotalAccepted,
						TotalAcceptedRaw:   q.Stats.TotalAcceptedRaw,
						TotalSubmission:    q.Stats.TotalSubmission,
						TotalSubmissionRaw: q.Stats.TotalSubmissionRaw,
						ACRate:             q.Stats.ACRate,
						Content:            content,
						Hints:              q.Hints,
					},
				)
			}
		}
		if len(questions) == 0 {
			return errors.New("no questions found")
		}

		switch flagFormat {
		default:
			outputHuman(questions, cmd.OutOrStdout())
		case "json":
			outputJson(questions, cmd.OutOrStdout())
		}

		return nil
	},
}

func outputHuman(qs []question, out io.Writer) {
	w := table.NewWriter()
	w.SetOutputMirror(out)
	w.SetStyle(table.StyleColoredDark)
	w.SetColumnConfigs(
		[]table.ColumnConfig{
			{
				Number:   2,
				WidthMax: 50,
			},
		},
	)
	for _, q := range qs {
		w.AppendRow(
			table.Row{fmt.Sprintf("%s. %s", q.FrontendId, q.Title)},
			table.RowConfig{AutoMerge: true, AutoMergeAlign: text.AlignLeft},
		)
		w.AppendRow(table.Row{"Slug", q.Slug})
		w.AppendRow(table.Row{"Difficulty", q.Difficulty})
		w.AppendRow(table.Row{"URL", q.Url})
		w.AppendRow(table.Row{"Tags", strings.Join(q.Tags, ", ")})
		w.AppendRow(table.Row{"Paid Only", q.IsPaidOnly})
		w.AppendRow(
			table.Row{
				"AC Rate",
				fmt.Sprintf("%s/%s %s", q.TotalAccepted, q.TotalSubmission, q.ACRate),
			},
		)
		if q.Content != "" {
			w.AppendRow(table.Row{"Content", q.Content})
		}
		for _, h := range q.Hints {
			w.AppendRow(table.Row{"Hint", h})
		}
		w.AppendSeparator()
	}
	w.Render()
}

func outputJson(qs []question, out io.Writer) {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	_ = enc.Encode(qs)
}
