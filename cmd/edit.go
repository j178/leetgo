package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/editor"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
)

var editCmd = &cobra.Command{
	Use:     "edit qid",
	Short:   "Open solution in editor",
	Aliases: []string{"e"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		c := leetcode.NewClient()
		qs, err := leetcode.ParseQID(args[0], c)
		if err != nil {
			return err
		}
		if len(qs) > 1 {
			return fmt.Errorf("multiple questions found")
		}
		_, err = lang.GetGenerator(cfg.Code.Lang)
		if err != nil {
			return err
		}
		result, err := lang.GeneratePathsOnly(qs[0])
		if err != nil {
			return err
		}
		return editor.Open(result.Files)
	},
}
