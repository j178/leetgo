package cmd

import (
	"text/template"

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
			q, _ := leetcode.Question(s, c)
			err := t.Execute(cmd.OutOrStdout(), &q)
			if err != nil {
				return err
			}
		}
		return nil
	},
}
