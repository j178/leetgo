package cmd

import (
	"fmt"
	"time"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use:   "submit SLUG_OR_ID",
	Short: "Submit solution",
	RunE: func(cmd *cobra.Command, args []string) error {
		cred := leetcode.CredentialsFromConfig()
		// cred := leetcode.NonAuth()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		fmt.Println(c.GetUserStatus())
		q, err := leetcode.Question("first-letter-to-appear-twice", c)
		if err != nil {
			return err
		}
		code := `
func repeatedCharacter(s string) byte {
    occu := [26]bool{}
    for _, c := range s {
        if occu[c-'a']{
            return byte(c)
        }
        occu[c-'a'] = true
    }
    return 'a'
}`
		r, err := c.InterpretSolution(q, config.Get().Gen, code, "\"abccbaacz\"")
		if err != nil {
			return err
		}
		fmt.Println(r)
		for i := 0; i < 10; i++ {
			checkResult, err := c.CheckSubmissionResult("runcode_1672574675.3374681_BNu5TMLwRs")
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", checkResult)
			time.Sleep(1 * time.Second)
		}
		return nil
	},
}
