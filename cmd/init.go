package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
	"github.com/spf13/cobra"
)

var (
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
		if initTemplate != "us" && initTemplate != "cn" {
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
		err = createQuestionDB()
		if err != nil {
			return err
		}
		err = createLibrary(dir)
		return err
	},
}

func init() {
	initCmd.Flags().StringVarP(&initTemplate, "template", "t", "us", "template to use, cn or us")
	_ = initCmd.MarkFlagRequired("template")
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
	hclog.L().Info("config dir created", "dir", dir)
	return nil
}

func createConfigFiles(dir string) error {
	cfg := config.Get()
	var site config.LeetcodeSite
	var language config.Language
	if initTemplate == "us" {
		site = config.LeetCodeUS
		language = config.EN
	} else {
		site = config.LeetCodeCN
		language = config.ZH
	}

	globalFile := cfg.GlobalConfigFile()
	if !utils.IsExist(globalFile) {
		f, err := os.Create(globalFile)
		if err != nil {
			return err
		}

		cfg.LeetCode.Site = site
		cfg.Language = language

		err = cfg.Write(f)
		if err != nil {
			return err
		}
		hclog.L().Info("global config file created", "file", globalFile)
	}

	projectFile := filepath.Join(dir, cfg.ProjectConfigFilename())
	f, err := os.Create(projectFile)
	if err != nil {
		return err
	}
	tmpl := `# leetgo project level config, global config is at %s
language: %s
code:
  lang: %s
editor:
  use: none
leetcode:
  site: %s
  credentials:
    read_from_browser: %s
`
	_, _ = f.WriteString(
		fmt.Sprintf(
			tmpl,
			globalFile,
			language,
			cfg.Code.Lang,
			site,
			cfg.LeetCode.Credentials.ReadFromBrowser,
		),
	)
	hclog.L().Info("project config file created", "file", projectFile)

	return nil
}

func createQuestionDB() error {
	if utils.IsExist(config.Get().LeetCodeCacheFile()) {
		return nil
	}
	c := leetcode.NewClient()
	err := leetcode.GetCache().Update(c)
	if err != nil {
		return err
	}
	return nil
}

func createLibrary(dir string) error {
	cfg := config.Get()
	gen := lang.GetGenerator(cfg.Code.Lang)
	if gen == nil {
		return fmt.Errorf("language %s is not supported yet, welcome to send a PR", cfg.Code.Lang)
	}
	if gen.CheckLibrary(dir) {
		return nil
	}
	return gen.GenerateLibrary(dir)
}
