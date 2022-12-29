package config

import (
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
	CN       bool           `json:"cn" yaml:"cn"`
	LeetCode LeetCodeConfig `json:"leetcode" yaml:"leetcode"`
	Go       GoConfig       `json:"go" yaml:"go"`
	Python   PythonConfig   `json:"python" yaml:"python"`
	// Add more languages here
	dir string
}

type PythonConfig struct {
	Enable bool   `json:"enable" yaml:"enable"`
	OutDir string `json:"out_dir" yaml:"out_dir"`
}

type GoConfig struct {
	Enable           bool   `json:"enable" yaml:"enable"`
	OutDir           string `json:"out_dir" yaml:"out_dir"`
	SeparatePackage  bool   `json:"separate_package" yaml:"separate_package"`
	FilenameTemplate string `json:"filename_template" yaml:"filename_template"`
}

type LeetCodeConfig struct {
	Site Site `json:"site" yaml:"site"`
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
