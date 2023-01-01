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

var initCmd = &cobra.Command{
	Use:     "init DIR",
	Short:   "Init a leetcode workspace",
	Example: "leetgo init . --gen go",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
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
	globalFile := cfg.GlobalConfigFile()
	if !utils.IsExist(globalFile) {
		f, err := os.Create(globalFile)
		if err != nil {
			return err
		}
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
	tmpl := `# leetgo project level config
gen: %s
leetcode:
  site: %s
  credentials:
    read_from_browser: %s
`
	_, _ = f.WriteString(fmt.Sprintf(tmpl, cfg.Gen, cfg.LeetCode.Site, cfg.LeetCode.Credentials.ReadFromBrowser))
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
	gen := lang.GetGenerator(cfg.Gen)
	if gen == nil {
		return fmt.Errorf("language %s is not supported yet, welcome to send a PR", cfg.Gen)
	}
	if gen.CheckLibrary() {
		return nil
	}
	return gen.GenerateLibrary()
}
