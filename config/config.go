package config

import (
	"path/filepath"

	"github.com/j178/leetgo/lang"
	"github.com/mitchellh/go-homedir"
)

const (
	cmdName           = "leetgo"
	configFile        = "config.yml"
	leetcodeCacheFile = "cache/leetcode-questions.json"
)

var cfg *Config

type Config struct {
	Cn       bool              `json:"cn" yaml:"cn"`
	LeetCode LeetCodeConfig    `json:"leetcode" yaml:"leetcode"`
	Go       lang.GoConfig     `json:"go" yaml:"go"`
	Python   lang.PythonConfig `json:"python" yaml:"python"`
	// Add more languages here
	dir string
}

type LeetCodeConfig struct {
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
		dir:      configDir,
		Cn:       true,
		LeetCode: LeetCodeConfig{},
		Go: lang.GoConfig{
			Enable:           false,
			SeparatePackage:  true,
			FilenameTemplate: ``,
		},
		Python: lang.PythonConfig{
			Enable: false,
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
