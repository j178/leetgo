package cmd

import (
	"bytes"

	"github.com/goccy/go-json"
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:    "inspect",
	Short:  "Inspect LeetCode API, developer only",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient()
		resp, err := c.Inspect(args[0])
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "  ")
		_ = enc.Encode(resp)
		cmd.Print(buf.String())
		return nil
	},
}
