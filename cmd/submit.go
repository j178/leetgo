package cmd

import (
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use:     "submit qid",
	Short:   "Submit solution",
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		gen := lang.GetGenerator(cfg.Code.Lang)
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		qs, err := leetcode.ParseQID(args[0], c)
		if err != nil {
			return err
		}

		limitC := make(chan struct{})
		defer func() { close(limitC) }()
		go func() {
			for {
				if _, ok := <-limitC; !ok {
					return
				}
				time.Sleep(2 * time.Second)
			}
		}()

		for _, q := range qs {
			limitC <- struct{}{}
			result, err := submitSolution(q, c, gen)
			if err != nil {
				hclog.L().Error("failed to submit solution", "question", q.TitleSlug, "err", err)
				continue
			}
			cmd.Print(result.Display(qs[0]))
		}

		return nil
	},
}
