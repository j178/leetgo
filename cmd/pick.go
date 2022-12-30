package cmd

import (
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var pickCmd = &cobra.Command{
	Use:     "pick [SLUG_OR_ID...]",
	Short:   "Generate a new question",
	Example: "leetgo pick 450\nleetgo pick two-sum",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient()
		if len(args) == 0 {
			// 	TODO Start tea TUI to pick a question
			args = append(args, "two-sum")
		}
		for _, p := range args {
			q, err := leetcode.Question(p, c)
			if err != nil {
				cmd.Printf("Failed to get question %s: %v\n", p, err)
			}
			files, err := lang.Generate(q)
			if err != nil {
				cmd.Printf("Failed to generate %s: %v\n", q.TitleSlug, err)
				continue
			}
			for _, f := range files {
				cmd.Printf("Generated %s\n", f)
			}
		}
		return nil
	},
}
