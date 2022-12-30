package cmd

import (
	"text/template"

	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

const tmpl = `{{ .QuestionFrontendId }}. {{ .Title }}
Slug: {{ .TitleSlug }}
Difficulty: {{ .Difficulty }}
URL: {{ .Url }}
{{ range $i, $t := .Hints }}Hint: {{ $t }}{{ end }}
`

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show question info",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient()
		t := template.Must(template.New("").Parse(tmpl))
		for _, s := range args {
			q, err := leetcode.Question(s, c)
			if err != nil {
				hclog.L().Error("failed to get question", "slug", s, "err", err)
				continue
			}
			err = t.Execute(cmd.OutOrStdout(), &q)
			if err != nil {
				return err
			}
		}
		return nil
	},
}
