package config

import (
	"fmt"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

const (
	cmdName           = "leetgo"
	configFile        = "config.yml"
	leetcodeCacheFile = "cache/leetcode-questions.json"
)

var cfg *Config

type Site string

const (
	LeetCodeCN Site = "https://leetcode.cn"
	LeetCodeUS Site = "https://leetcode.com"
)

type Config struct {
	CN       bool           `yaml:"cn"`
	LeetCode LeetCodeConfig `yaml:"leetcode"`
	Go       GoConfig       `yaml:"go"`
	Python   PythonConfig   `yaml:"python"`
	// Add more languages here
	dir string
}

type PythonConfig struct {
	Enable bool   `yaml:"enable"`
	OutDir string `yaml:"out_dir"`
}

type GoConfig struct {
	Enable           bool   `yaml:"enable"`
	OutDir           string `yaml:"out_dir"`
	SeparatePackage  bool   `yaml:"separate_package"`
	FilenameTemplate string `yaml:"filename_template"`
}

type LeetCodeConfig struct {
	Site Site `yaml:"site"`
}

func (c Config) ConfigDir() string {
	return c.dir
}

func (c Config) ConfigFile() string {
	return filepath.Join(c.dir, configFile)
}

func (c Config) LeetCodeCacheFile() string {
	return filepath.Join(c.dir, leetcodeCacheFile)
}

func Default() Config {
	home, _ := homedir.Dir()
	configDir := filepath.Join(home, ".config", cmdName)
	return Config{
		dir: configDir,
		CN:  true,
		LeetCode: LeetCodeConfig{
			Site: LeetCodeCN,
		},
		Go: GoConfig{
			Enable:           false,
			OutDir:           "go",
			SeparatePackage:  true,
			FilenameTemplate: ``,
		},
		Python: PythonConfig{
			Enable: false,
			OutDir: "python",
		},
		// Add more languages here
	}
}

func Get() Config {
	if cfg == nil {
		return Default()
	}
	return *cfg
}

func Init(c Config) {
	cfg = &c
}

func Verify(c Config) error {
	if c.LeetCode.Site != LeetCodeCN && c.LeetCode.Site != LeetCodeUS {
		return fmt.Errorf("invalid site: %s", c.LeetCode.Site)
	}

	return nil
}
