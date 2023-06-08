package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/log"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

// Use ChatGPT API to fix solution code

var fixCmd = &cobra.Command{
	Use:   "fix qid",
	Short: "Use ChatGPT API to fix your solution code (just for fun)",
	Long: `Use ChatGPT API to fix your solution code.
Set OPENAI_API_KEY environment variable to your OpenAI API key before using this command.`,
	Example: `leetgo fix 429`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient(leetcode.ReadCredentials())
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

Please identify any issues or inefficiencies in my code and to help me fix or improve it. DO NOT rewrite it.
I want you to only reply with pure code without <code> or markdown tags, and nothing else. DO NOT write explanations.
`

var errNoFix = errors.New("no fix found")

func askOpenAI(cmd *cobra.Command, q *leetcode.QuestionData, code string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("missing OPENAI_API_KEY environment variable, you can find or create your API key here: https://platform.openai.com/account/api-keys")
	}
	baseURI := os.Getenv("OPENAI_API_ENDPOINT")
	config := openai.DefaultConfig(apiKey)
	if baseURI != "" {
		config.BaseURL = baseURI
	}
	client := openai.NewClientWithConfig(config)
	prompt := fmt.Sprintf(
		fixPrompt,
		q.Title,
		q.GetFormattedContent(),
		code,
	)
	log.Debug("requesting openai", "prompt", prompt)
	spin := newSpinner(cmd.OutOrStdout())
	spin.Suffix = " Waiting for OpenAI..."
	spin.Start()
	defer spin.Stop()

	ctx := context.Background()
	resp, err := client.CreateChatCompletion(
		ctx, openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{Role: "system", Content: "Help solve LeetCode questions and fix the code"},
				{Role: "user", Content: prompt},
			},
			MaxTokens:   1000,
			Temperature: 0,
		},
	)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errNoFix
	}
	log.Debug("got response from openai", "response", resp.Choices)
	text := resp.Choices[0].Message.Content
	text = utils.EnsureTrailingNewline(text)
	return text, nil
}
