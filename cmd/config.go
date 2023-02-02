package cmd

import (
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/j178/leetgo/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show configurations",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()
		cwd, _ := os.Getwd()
		projectConfig, err := os.ReadFile(cfg.ProjectConfigFile())
		if err != nil {
			projectConfig = []byte("No project config file found")
		}
		cmd.Println("Leetgo version info  :")
		cmd.Println("```")
		cmd.Println(buildVersion())
		cmd.Println("```")
		cmd.Println("Global config dir    :", cfg.ConfigDir())
		cmd.Println("Global config file   :", cfg.GlobalConfigFile())
		cmd.Println("Project root         :", cfg.ProjectRoot())
		cmd.Println("Working dir          :", cwd)
		cmd.Println("Project config file  :", cfg.ProjectConfigFile())
		cmd.Println("Project configuration:")
		cmd.Println("```yaml")
		cmd.Println(string(projectConfig))
		cmd.Println("```")
		cmd.Println("Full configuration   :")
		cmd.Println("```yaml")
		_ = cfg.Write(cmd.OutOrStdout(), false)
		cmd.Println("```")
	},
}

var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt a sensitive string to be used in config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt := &survey.Password{
			Message: "Enter the string to be encrypted",
		}
		var input string
		err := survey.AskOne(prompt, &input)
		if err != nil {
			return err
		}
		encrypted, err := config.Encrypt(input)
		if err != nil {
			return err
		}
		cmd.Println()
		cmd.Println(encrypted)
		return nil
	},
}

func init() {
	configCmd.AddCommand(encryptCmd)
}
