package cmd

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/hako/durafmt"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/editor"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
)

var (
	// https://robotmoon.com/256-colors/
	contestTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("47")).Bold(true)
	nameStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("47"))
	timeStyle         = lipgloss.NewStyle().Faint(true)
	checkDuration     = 10 * time.Second
	openInBrowser     bool
)

func selectUpcomingContest(c leetcode.Client, registeredOnly bool) (string, error) {
	contestList, err := c.GetUpcomingContests()
	if err != nil {
		return "", err
	}
	if registeredOnly {
		var list []*leetcode.Contest
		for _, ct := range contestList {
			if ct.Registered {
				list = append(list, ct)
			}
		}
		contestList = list
	}

	if len(contestList) == 0 {
		msg := "no upcoming contest"
		if registeredOnly {
			msg = "no registered contest"
		}
		return "", errors.New(msg)
	}

	contestNames := make([]string, len(contestList))
	for i, ct := range contestList {
		mark := " "
		if ct.Registered {
			mark = "√"
		}
		contestNames[i] = fmt.Sprintf(
			"%s %s at %s",
			mark,
			ct.Title,
			time.Unix(ct.StartTime, 0).Format("2006/01/02 15:04:05"),
		)
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

func waitContestStart(cmd *cobra.Command, ct *leetcode.Contest) error {
	if ct.HasStarted() {
		return nil
	}

	var mu sync.Mutex
	spin := newSpinner(cmd.ErrOrStderr())
	spin.PreUpdate = func(s *spinner.Spinner) {
		mu.Lock()
		defer mu.Unlock()
		s.Suffix = fmt.Sprintf(
			" %s begins in %s, waiting...",
			contestTitleStyle.Render(ct.Title),
			timeStyle.Render(durafmt.Parse(ct.TimeTillStart()).LimitFirstN(2).String()),
		)
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
	Use:   "contest [qid]",
	Short: "Generate contest questions",
	Example: `leetgo contest
leetgo contest w330
leetgo contest left w330
`,
	Aliases: []string{"c"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cred, err := leetcode.ReadCredentials()
		if err != nil {
			return err
		}
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		cfg := config.Get()

		var qid string
		if len(args) == 0 {
			qid, err = selectUpcomingContest(c, false)
			if err != nil {
				return err
			}
		} else {
			qid = args[0]
		}
		if slash := strings.Index(qid, "/"); slash > 0 && slash != len(qid)-1 {
			log.Warn("ignore question ID part in qid", "qid", qid)
		}
		if !strings.Contains(qid, "/") {
			qid += "/"
		}

		contest, _, err := leetcode.ParseContestQID(qid, c, false)
		if err != nil {
			return err
		}
		user, err := c.GetUserStatus()
		if err != nil {
			user = &leetcode.UserStatus{}
		}

		if !contest.HasFinished() && !contest.Registered {
			register := true
			if !viper.GetBool("yes") {
				prompt := survey.Confirm{
					Message: fmt.Sprintf(
						"Register for %s as %s?",
						contestTitleStyle.Render(contest.Title),
						nameStyle.Render(user.Whoami(c)),
					),
				}
				err := survey.AskOne(&prompt, &register)
				if err != nil {
					return err
				}
			}
			if register {
				err = c.RegisterContest(contest.TitleSlug)
				if err != nil {
					return err
				}
				log.Info("registered", "contest", contest.Title, "user", user.Whoami(c))
			} else {
				return nil
			}
		}

		err = waitContestStart(cmd, contest)
		if err != nil {
			return err
		}

		generated, err := lang.GenerateContest(contest)
		if err != nil {
			return err
		}

		isSet := cmd.Flags().Lookup("browser").Changed
		if (isSet && openInBrowser) || (!isSet && cfg.Contest.OpenInBrowser) {
			for _, r := range generated {
				_ = browser.OpenURL(r.Question.ContestUrl())
			}
		}
		err = editor.Open(generated[0])
		return err
	},
}

var unregisterCmd = &cobra.Command{
	Use:     "left [qid]",
	Short:   "Unregister from contest",
	Aliases: []string{"un", "unregister"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cred, err := leetcode.ReadCredentials()
		if err != nil {
			return err
		}
		c := leetcode.NewClient(leetcode.WithCredentials(cred))

		var qid string
		if len(args) == 0 {
			qid, err = selectUpcomingContest(c, true)
			if err != nil {
				return err
			}
		} else {
			qid = args[0]
		}
		if !strings.HasSuffix(qid, "/") {
			qid += "/"
		}

		contest, _, err := leetcode.ParseContestQID(qid, c, false)
		if err != nil {
			return err
		}
		if !contest.Registered {
			return fmt.Errorf("you are not registered for %s", contest.Title)
		}
		if contest.HasFinished() {
			return fmt.Errorf("contest %s has finished", contest.Title)
		}
		user, err := c.GetUserStatus()
		if err != nil {
			return err
		}
		unregister := true
		if !viper.GetBool("yes") {
			prompt := survey.Confirm{
				Message: fmt.Sprintf(
					"Unregister from %s as %s?",
					contestTitleStyle.Render(contest.Title),
					nameStyle.Render(user.Whoami(c)),
				),
			}
			err = survey.AskOne(&prompt, &unregister)
			if err != nil {
				return err
			}
		}
		if unregister {
			err = c.UnregisterContest(contest.TitleSlug)
			if err != nil {
				return err
			}
			log.Info("unregistered", "contest", contest.Title, "user", user.Whoami(c))
		}

		return nil
	},
}

func init() {
	contestCmd.Flags().BoolVarP(&openInBrowser, "browser", "b", false, "open question page in browser")
	contestCmd.AddCommand(unregisterCmd)
}
