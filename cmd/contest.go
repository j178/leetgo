package cmd

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
)

var contestCmd = &cobra.Command{
	Use:     "contest",
	Short:   "Generate contest questions",
	Aliases: []string{"c"},
	RunE: func(cmd *cobra.Command, args []string) error {
		spin := spinner.New(
			spinner.CharSets[9],
			250*time.Millisecond,
			spinner.WithHiddenCursor(true),
			spinner.WithSuffix("Waiting for contest..."),
		)
		spin.Start()
		defer spin.Stop()
		time.Sleep(10 * time.Second)
		return nil
	},
}
