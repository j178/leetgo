package cmd

import "github.com/spf13/cobra"

var contestCmd = &cobra.Command{
	Use:   "contest",
	Short: "Generate contest questions",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
