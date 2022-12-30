package cmd

import (
	"github.com/hashicorp/go-hclog"
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
				hclog.L().Error("failed to get question", "question", p, "error", err)
				continue
			}
			files, err := lang.Generate(q)
			if err != nil {
				hclog.L().Error("failed to generate", "question", p, "error", err)
				continue
			}
			for _, f := range files {
				hclog.L().Info("generated", "files", f)
			}
		}
		return nil
	},
}
