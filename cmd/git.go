package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
)

var gitCmd = &cobra.Command{
	Use:    "git",
	Hidden: true,
	Short:  "Git related commands",
}

var gitPushCmd = &cobra.Command{
	Use:   "push qid",
	Short: "Add, commit and push your solution to remote repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient(leetcode.ReadCredentials())
		qid := args[0]
		qs, err := leetcode.ParseQID(qid, c)
		if err != nil {
			return err
		}
		if len(qs) > 1 {
			return fmt.Errorf("multiple questions found")
		}
		result, err := lang.GeneratePathsOnly(qs[0])
		if err != nil {
			return err
		}
		err = gitAddCommitPush(result)
		return err
	},
}

func init() {
	gitCmd.AddCommand(gitPushCmd)
}

func gitAddCommitPush(genResult *lang.GenerateResult) error {
	files := make([]string, 0, len(genResult.Files))
	for _, f := range genResult.Files {
		files = append(files, f.GetPath())
	}
	err := runCmd("git", "add", files...)
	if err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	var msg string
	prompt := &survey.Input{
		Message: "Commit message",
		Default: fmt.Sprintf(
			"Add solution for %s.",
			genResult.Question.TitleSlug,
		),
	}
	err = survey.AskOne(prompt, &msg)
	if err != nil {
		return fmt.Errorf("git commit message: %w", err)
	}

	msg = stripComments(msg)
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return errors.New("git: empty commit message")
	}
	err = runCmd("git", "commit", "-m", msg)
	if err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	err = runCmd("git", "push")
	if err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	return nil
}

func stripComments(s string) string {
	lines := strings.Split(s, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

func runCmd(command string, subcommand string, args ...string) error {
	cmd := exec.Command(command, subcommand)
	cmd.Args = append(cmd.Args, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
