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
		err = createConfigFiles(dir)
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
	initCmd.Flags().StringVarP(&initTemplate, "template", "t", "us", "template to use, cn or us")
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite global config file if exists")
}

func createConfigDir() error {
	dir := config.Get().ConfigDir()
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

func createConfigFiles(dir string) error {
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

	globalFile := cfg.GlobalConfigFile()
	if force || !utils.IsExist(globalFile) {
		f, err := os.Create(globalFile)
		if err != nil {
			return err
		}

		author := defaultUser()
		cfg.LeetCode.Site = site
		cfg.Language = language
		cfg.Author = author

		err = cfg.Write(f, true)
		if err != nil {
			return err
		}
		log.Info("global config file created", "file", globalFile)
	}

	projectFile := filepath.Join(dir, constants.ProjectConfigFilename)
	f, err := os.Create(projectFile)
	if err != nil {
		return err
	}

	tmpl := `# leetgo project level config, global config is at %s
# for more details, please refer to https://github.com/j178/leetgo
language: %s
code:
  lang: %s
leetcode:
  site: %s
#  credentials:
#    from: browser
#editor:
#  use: none
`
	_, _ = f.WriteString(
		fmt.Sprintf(
			tmpl,
			globalFile,
			language,
			cfg.Code.Lang,
			site,
		),
	)
	log.Info("project config file created", "file", projectFile)

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
