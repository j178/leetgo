package cmd

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	colorGreen    = color.New(color.FgHiGreen, color.Bold)
	checkDuration = 10 * time.Second
)

func selectUpcomingContest(c leetcode.Client) (string, error) {
	contestList, err := c.GetUpcomingContests()
	if err != nil {
		return "", err
	}
	if len(contestList) == 0 {
		return "", fmt.Errorf("no upcoming contest")
	}
	contestNames := make([]string, len(contestList))
	for i, ct := range contestList {
		mark := " "
		if ct.Registered {
			mark = colorGreen.Sprint("√")
		}
		contestNames[i] = mark + " " + ct.Title
	}
	var idx int
	prompt := &survey.Select{
		Message: "Select a contest:",
		Options: contestNames,
	}
	err = survey.AskOne(prompt, &idx)
	if err != nil {
		return "", err
	}
	return contestList[idx].TitleSlug, nil
}

func waitContestStart(ct *leetcode.Contest) error {
	var mu sync.Mutex

	spin := spinner.New(spinner.CharSets[9], 250*time.Millisecond)
	spin.PreUpdate = func(s *spinner.Spinner) {
		mu.Lock()
		defer mu.Unlock()
		t := ct.TimeTillStart().Round(time.Second)
		s.Suffix = fmt.Sprintf("  %s begins in %s, waiting...", ct.Title, t)
	}
	spin.Start()
	defer spin.Stop()

	for {
		if ct.HasStarted() {
			return nil
		}
		wait := ct.TimeTillStart()
		if wait > checkDuration {
			wait = checkDuration
		}
		time.Sleep(wait)

		mu.Lock()
		err := ct.Refresh()
		mu.Unlock()
		if err != nil {
			return err
		}
	}
}

var contestCmd = &cobra.Command{
	Use:     "contest [qid]",
	Short:   "Generate contest questions",
	Aliases: []string{"c"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		cfg := config.Get()

		var contestSlug string
		var err error
		if len(args) == 0 {
			contestSlug, err = selectUpcomingContest(c)
			if err != nil {
				return err
			}
		} else {
			contestSlug = args[0]
		}
		if !strings.HasSuffix(contestSlug, "/") {
			contestSlug += "/"
		}

		contest, _, err := leetcode.ParseContestQID(contestSlug, c, false)
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

		err = waitContestStart(contest)
		if err != nil {
			return err
		}

		qs, err := contest.GetAllQuestions()
		if err != nil {
			return err
		}
		var generated []*lang.GenerateResult
		for _, q := range qs {
			result, err := lang.Generate(q)
			if err != nil {
				hclog.L().Error("failed to generate", "question", q.TitleSlug, "err", err)
				continue
			}
			generated = append(generated, result)
		}
		if len(generated) == 0 {
			return fmt.Errorf("no question generated")
		}
		if cfg.Contest.OpenInBrowser {
			for _, r := range generated {
				_ = browser.OpenURL(r.Question.ContestUrl())
			}
		}

		return nil
	},
}

var unregisterCmd = &cobra.Command{
	Use:     "left [qid]",
	Short:   "Unregister from contest",
	Aliases: []string{"un", "unregister"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))

		var contestSlug string
		var err error
		if len(args) == 0 {
			contestSlug, err = selectUpcomingContest(c)
			if err != nil {
				return err
			}
		} else {
			contestSlug = args[0]
		}
		if !strings.HasSuffix(contestSlug, "/") {
			contestSlug += "/"
		}

		contest, _, err := leetcode.ParseContestQID(contestSlug, c, false)
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
