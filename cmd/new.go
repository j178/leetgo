package cmd

import (
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:     "new SLUG_OR_ID...",
	Short:   "Generate a new question",
	Example: "leetgo new 450 --go\nleetgo new two-sum --go",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient()
		gen := lang.NewMultiGenerator()
		for _, p := range args {
			q, err := leetcode.Question(p, c)
			if err != nil {
				cmd.Printf("Failed to get question %s: %v\n", p, err)
			}
			files, err := gen.Generate(q)
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

func init() {
	addLangFlags(newCmd)
}
