package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PullRequestInc/go-gpt3"
	"github.com/charmbracelet/glamour"
	"github.com/hashicorp/go-hclog"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

// Use OpenAI GPT-3 API to fix solution code

var fixCmd = &cobra.Command{
	Use:   "fix qid",
	Short: "Use OpenAI GPT-3 API to fix your solution code (just for fun)",
	Long: `Use OpenAI GPT-3 API to fix your solution code.
Set OPENAI_API_KEY environment variable to your OpenAI API key before using this command.`,
	Example: `leetgo fix 429`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient(leetcode.WithCredentials(leetcode.CredentialsFromConfig()))
		qs, err := leetcode.ParseQID(args[0], c)
		if err != nil {
			return err
		}
		if len(qs) > 1 {
			return fmt.Errorf("multiple questions found")
		}
		q := qs[0]
		err = q.Fulfill()
		if err != nil {
			return err
		}

		code, err := lang.GetSolutionCode(q)
		if err != nil {
			return err
		}

		fixedCode, err := askOpenAI(cmd, q, code)
		if err != nil {
			return err
		}

		output := "# Here is the fix from OpenAI GPT-3 API\n"
		edits := myers.ComputeEdits("", code, fixedCode)
		diff := gotextdiff.ToUnified("original", "AI fixed", code, edits)
		output += "```diff\n" + fmt.Sprint(diff) + "\n```\n"
		output, err = glamour.Render(output, "dark")
		if err != nil {
			return err
		}
		cmd.Println(output)

		accept := true
		if !viper.GetBool("yes") {
			err = survey.AskOne(
				&survey.Confirm{
					Message: "Do you want to accept the fix?",
				}, &accept,
			)
			if err != nil {
				return err
			}
		}
		if accept {
			err = lang.UpdateSolutionCode(q, fixedCode)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

const fixPrompt = `Given a LeetCode problem %s, the problem description below is wrapped in <question> and </question> tags. The solution code is wrapped in <code> and </code> tags:
<question>
%s
</question>

I have written the following solution:
%s

Please identify any issues or inefficiencies in my code and to help me fix or improve it.
I want you to only reply with pure code without <code> tags, and nothing else. Do not write explanations.
`

var errNoFix = errors.New("no fix found")

func askOpenAI(cmd *cobra.Command, q *leetcode.QuestionData, code string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("missing OPENAI_API_KEY environment variable")
	}
	client := gpt3.NewClient(apiKey)
	prompt := fmt.Sprintf(
		fixPrompt,
		q.Title,
		q.GetFormattedContent(),
		code,
	)
	hclog.L().Debug("requesting openai", "prompt", prompt)
	spin := newSpinner(cmd.OutOrStdout())
	spin.Suffix = " Waiting for OpenAI..."
	spin.Start()
	defer spin.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	resp, err := client.CompletionWithEngine(
		ctx, gpt3.TextDavinci003Engine, gpt3.CompletionRequest{
			Prompt:      []string{prompt},
			MaxTokens:   gpt3.IntPtr(3000),
			Temperature: gpt3.Float32Ptr(0),
		},
	)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errNoFix
	}
	hclog.L().Debug("got response from openai", "response", resp.Choices)
	text := resp.Choices[0].Text
	text = utils.EnsureTrailingNewline(text)
	return text, nil
}
