package cmd

import "github.com/spf13/cobra"

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run question test cases",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	testCmd.Flags().BoolP("submit", "s", false, "auto submit if all tests passed")
}
