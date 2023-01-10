package cmd

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var contestCmd = &cobra.Command{
	Use:     "contest [qid]",
	Short:   "Generate contest questions",
	Aliases: []string{"c"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var contestSlug string
		if len(args) == 0 {
			// get upcoming contest
			// select to register / unregister
			// register then wait for contest to start
			contestSlug = "weekly-contest-328"
		} else {
			contestSlug = args[0]
		}
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		contest, err := c.GetContest(contestSlug)
		if err != nil {
			return err
		}

		if !contest.HasFinished() && !contest.Registered {
			register := true
			if !viper.GetBool("yes") {
				prompt := survey.Confirm{
					Message: fmt.Sprintf("Register for %s?", contest.Title),
				}
				err := survey.AskOne(&prompt, &register)
				if err != nil {
					return err
				}
			}
			if register {
				hclog.L().Info("registering for contest", "contest", contest.Title)
				err = c.RegisterContest(contestSlug)
				if err != nil {
					return err
				}
				hclog.L().Info("registered", "contest", contest.Title)
			} else {
				return nil
			}
		}

		_, _ = contest.GetAllQuestions()
		fmt.Printf("%+v", contest)

		return nil
	},
}

var unregisterCmd = &cobra.Command{
	Use:   "left",
	Short: "Unregister from contest",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// list joined contests to select
		// confirm unregister
		// unregister
		var contestSlug string
		if len(args) == 0 {
			// get upcoming contest
			// select to register / unregister
			// register then wait for contest to start
			contestSlug = "weekly-contest-328"
		} else {
			contestSlug = args[0]
		}
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		contest, err := c.GetContest(contestSlug)
		if err != nil {
			return err
		}
		if !contest.Registered {
			return fmt.Errorf("you are not registered for %s", contest.Title)
		}
		if contest.HasFinished() {
			return fmt.Errorf("contest %s has finished", contest.Title)
		}
		unregister := true
		if !viper.GetBool("yes") {
			prompt := survey.Confirm{
				Message: fmt.Sprintf("Unregister from %s?", contest.Title),
			}
			err = survey.AskOne(&prompt, &unregister)
			if err != nil {
				return err
			}
		}
		if unregister {
			hclog.L().Info("unregistering from contest", "contest", contest.Title)
			err = c.UnregisterContest(contestSlug)
			if err != nil {
				return err
			}
			hclog.L().Info("unregistered", "contest", contest.Title)
		}

		return nil
	},
}

func init() {
	contestCmd.AddCommand(unregisterCmd)
}
