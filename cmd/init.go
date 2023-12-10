package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/constants"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

var (
	force        bool
	initTemplate string
)

var initCmd = &cobra.Command{
	Use:     "init [DIR]",
	Short:   "Init a leetcode workspace",
	Example: "leetgo init -t us -l cpp",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}
		dir, _ = filepath.Abs(dir)
		if initTemplate != "" && initTemplate != "us" && initTemplate != "cn" {
			return fmt.Errorf("invalid template %s, only us or cn is supported", initTemplate)
		}
		err := utils.CreateIfNotExists(dir, true)
		if err != nil {
			return err
		}
		err = createConfigDir()
		if err != nil {
			return err
		}
		err = createConfigFile(dir)
		if err != nil {
			return err
		}
		if gitAvailable() && !isInsideGitRepo(dir) {
			_ = initGitRepo(dir)
		}
		err = createQuestionCache()
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	initCmd.Flags().StringVarP(&initTemplate, "template", "t", "", "template to use, cn or us")
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite config file if exists")

	_ = initCmd.RegisterFlagCompletionFunc(
		"template", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{"us", "cn"}, cobra.ShellCompDirectiveNoFileComp
		},
	)
}

func createConfigDir() error {
	dir := config.Get().HomeDir()
	if utils.IsExist(dir) {
		return nil
	}
	err := utils.MakeDir(dir)
	if err != nil {
		return err
	}
	log.Info("config dir created", "dir", dir)
	return nil
}

func createConfigFile(dir string) error {
	cfg := config.Get()
	site := cfg.LeetCode.Site
	language := cfg.Language
	if initTemplate == "us" {
		site = config.LeetCodeUS
		language = config.EN
	} else if initTemplate == "cn" {
		site = config.LeetCodeCN
		language = config.ZH
	}

	author := defaultUser()
	cfg.LeetCode.Site = site
	cfg.Language = language
	cfg.Author = author

	projectFile := filepath.Join(dir, constants.ConfigFilename)
	if utils.IsExist(projectFile) && !force {
		return fmt.Errorf("config file %s already exists, use -f to overwrite", utils.RelToCwd(projectFile))
	}

	f, err := os.Create(projectFile)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	_, _ = f.WriteString("# Leetgo configuration file, see more at https://github.com/j178/leetgo\n\n")
	_ = cfg.Write(f, true)
	log.Info("config file created", "file", utils.RelToCwd(projectFile))

	return nil
}

func createQuestionCache() error {
	c := leetcode.NewClient(leetcode.ReadCredentials())
	cache := leetcode.GetCache(c)
	if !cache.Outdated() {
		return nil
	}
	err := cache.Update()
	if err != nil {
		return err
	}
	return nil
}

func defaultUser() string {
	username := getGitUsername()
	if username != "" {
		return username
	}
	u, err := user.Current()
	if err == nil {
		return u.Username
	}
	username = os.Getenv("USER")
	if username != "" {
		return username
	}
	return "Bob"
}

func gitAvailable() bool {
	cmd := exec.Command("git", "--version")
	err := cmd.Run()
	return err == nil
}

func initGitRepo(dir string) error {
	cmd := exec.Command("git", "init", dir)
	return cmd.Run()
}

func isInsideGitRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree", dir)
	err := cmd.Run()
	return err == nil
}

func getGitUsername() string {
	cmd := exec.Command("git", "config", "user.name")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
