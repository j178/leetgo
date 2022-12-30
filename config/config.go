package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

const (
	CmdName           = "leetgo"
	globalConfigFile  = "config.yaml"
	projectConfigFile = CmdName + ".yaml"
	leetcodeCacheFile = "cache/leetcode-questions.json"
)

var (
	cfg   *Config
	Debug = os.Getenv("DEBUG") != ""
)

type (
	Site     string
	Language string
)

const (
	LeetCodeCN Site     = "https://leetcode.cn"
	LeetCodeUS Site     = "https://leetcode.com"
	ZH         Language = "zh"
	EN         Language = "en"
)

type Config struct {
	dir        string
	Gen        string         `yaml:"gen" mapstructure:"gen" comment:"Generate code for questions, go, python, ... (will be override by project config and flag --gen)"`
	Language   Language       `yaml:"language" mapstructure:"language" comment:"Language of the questions, zh or en"`
	LeetCode   LeetCodeConfig `yaml:"leetcode" mapstructure:"leetcode" comment:"LeetCode configuration"`
	Editor     Editor         `yaml:"editor" mapstructure:"editor"`
	Go         GoConfig       `yaml:"go" mapstructure:"go"`
	Python     baseLangConfig `yaml:"python" mapstructure:"python"`
	Cpp        baseLangConfig `yaml:"cpp" mapstructure:"cpp"`
	Java       baseLangConfig `yaml:"java" mapstructure:"java"`
	Rust       baseLangConfig `yaml:"rust" mapstructure:"rust"`
	C          baseLangConfig `yaml:"c" mapstructure:"c"`
	CSharp     baseLangConfig `yaml:"csharp" mapstructure:"csharp"`
	JavaScript baseLangConfig `yaml:"javascript" mapstructure:"javascript"`
	Ruby       baseLangConfig `yaml:"ruby" mapstructure:"ruby"`
	Swift      baseLangConfig `yaml:"swift" mapstructure:"swift"`
	Kotlin     baseLangConfig `yaml:"kotlin" mapstructure:"kotlin"`
	PHP        baseLangConfig `yaml:"php" mapstructure:"php"`
	// Add more languages here
}

type Editor struct {
}

type baseLangConfig struct {
	OutDir string `yaml:"out_dir" mapstructure:"out_dir"`
}

type GoConfig struct {
	baseLangConfig   `yaml:",inline" mapstructure:",squash"`
	SeparatePackage  bool   `yaml:"separate_package" mapstructure:"separate_package" comment:"Generate separate package for each question"`
	FilenameTemplate string `yaml:"filename_template" mapstructure:"filename_template" comment:"Filename template for Go files"`
}

type LeetCodeConfig struct {
	Site Site `yaml:"site" mapstructure:"site" comment:"LeetCode site, https://leetcode.com or https://leetcode.cn"`
}

func (c Config) ConfigDir() string {
	return c.dir
}

func (c Config) GlobalConfigFile() string {
	return filepath.Join(c.dir, globalConfigFile)
}

func (c Config) ProjectConfigFile() string {
	return projectConfigFile
}

func (c Config) LeetCodeCacheFile() string {
	return filepath.Join(c.dir, leetcodeCacheFile)
}

func (c Config) WriteTo(w io.Writer) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	node, _ := toYamlNode(c)
	err := enc.Encode(node)
	return err
}

func Default() Config {
	home, _ := homedir.Dir()
	configDir := filepath.Join(home, ".config", CmdName)
	return Config{
		dir:      configDir,
		Gen:      "go",
		Language: ZH,
		LeetCode: LeetCodeConfig{
			Site: LeetCodeCN,
		},
		Go: GoConfig{
			baseLangConfig:   baseLangConfig{OutDir: "go"},
			SeparatePackage:  true,
			FilenameTemplate: ``,
		},
		Python:     baseLangConfig{OutDir: "python"},
		Cpp:        baseLangConfig{OutDir: "cpp"},
		Java:       baseLangConfig{OutDir: "java"},
		Rust:       baseLangConfig{OutDir: "rust"},
		C:          baseLangConfig{OutDir: "c"},
		CSharp:     baseLangConfig{OutDir: "csharp"},
		JavaScript: baseLangConfig{OutDir: "javascript"},
		Ruby:       baseLangConfig{OutDir: "ruby"},
		Swift:      baseLangConfig{OutDir: "swift"},
		Kotlin:     baseLangConfig{OutDir: "kotlin"},
		PHP:        baseLangConfig{OutDir: "php"},
		// Add more languages here
	}
}

func Get() Config {
	if cfg == nil {
		return Default()
	}
	return *cfg
}

func Set(c Config) {
	cfg = &c
}

func Verify(c Config) error {
	if c.LeetCode.Site != LeetCodeCN && c.LeetCode.Site != LeetCodeUS {
		return fmt.Errorf("invalid site: %s", c.LeetCode.Site)
	}
	if c.Language != ZH && c.Language != EN {
		return fmt.Errorf("invalid language: %s", c.Language)
	}
	if c.Gen == "" {
		return fmt.Errorf("gen is empty")
	}
	return nil
}
