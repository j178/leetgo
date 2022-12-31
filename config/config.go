package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/j178/leetgo/utils"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

const (
	CmdName               = "leetgo"
	globalConfigFile      = "config.yaml"
	projectConfigFilename = CmdName + ".yaml"
	leetcodeCacheFile     = "cache/leetcode-questions.db"
	CodeBeginMark         = "Leetgo Code Begin"
	CodeEndMark           = "Leetgo Code End"
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
	dir         string
	projectRoot string
	Author      string         `yaml:"author" mapstructure:"author" comment:"Your name"`
	Gen         string         `yaml:"gen" mapstructure:"gen" comment:"Generate code for questions, go, python, ... (will be override by project config and flag --gen)"`
	Language    Language       `yaml:"language" mapstructure:"language" comment:"Language of the questions, zh or en"`
	LeetCode    LeetCodeConfig `yaml:"leetcode" mapstructure:"leetcode" comment:"LeetCode configuration"`
	Contest     ContestConfig  `yaml:"contest" mapstructure:"contest"`
	Editor      Editor         `yaml:"editor" mapstructure:"editor" comment:"The editor to open generated files"`
	Cache       string         `yaml:"cache" mapstructure:"cache" comment:"Cache type, json or sqlite"`
	Go          GoConfig       `yaml:"go" mapstructure:"go"`
	Python      BaseLangConfig `yaml:"python" mapstructure:"python"`
	Cpp         BaseLangConfig `yaml:"cpp" mapstructure:"cpp"`
	Java        BaseLangConfig `yaml:"java" mapstructure:"java"`
	Rust        BaseLangConfig `yaml:"rust" mapstructure:"rust"`
	C           BaseLangConfig `yaml:"c" mapstructure:"c"`
	CSharp      BaseLangConfig `yaml:"csharp" mapstructure:"csharp"`
	JavaScript  BaseLangConfig `yaml:"javascript" mapstructure:"javascript"`
	Ruby        BaseLangConfig `yaml:"ruby" mapstructure:"ruby"`
	Swift       BaseLangConfig `yaml:"swift" mapstructure:"swift"`
	Kotlin      BaseLangConfig `yaml:"kotlin" mapstructure:"kotlin"`
	PHP         BaseLangConfig `yaml:"php" mapstructure:"php"`
	// Add more languages here
}

type ContestConfig struct {
	OutDir string `yaml:"out_dir" mapstructure:"out_dir" comment:"Base dir to put generated contest questions"`
}

type Editor struct {
	Command string   `yaml:"command" mapstructure:"command"`
	Args    []string `yaml:"args" mapstructure:"args"`
}

type BaseLangConfig struct {
	OutDir string `yaml:"out_dir" mapstructure:"out_dir"`
}

type GoConfig struct {
	BaseLangConfig   `yaml:",inline" mapstructure:",squash"`
	SeparatePackage  bool   `yaml:"separate_package" mapstructure:"separate_package" comment:"Generate separate package for each question"`
	FilenameTemplate string `yaml:"filename_template" mapstructure:"filename_template" comment:"Filename template for Go files"`
}

type Credentials struct {
	ReadFromBrowser string `yaml:"read_from_browser" mapstructure:"read_from_browser" comment:"Read leetcode cookie from browser, currently only chrome is supported."`
	Session         string `yaml:"session,omitempty" mapstructure:"session" comment:"LeetCode cookie: LEETCODE_SESSION"`
	CsrfToken       string `yaml:"csrf_token,omitempty" mapstructure:"csrf_token" comment:"LeetCode cookie: csrftoken"`
	Username        string `yaml:"username,omitempty" mapstructure:"username" comment:"LeetCode username"`
	Password        string `yaml:"password,omitempty" mapstructure:"password" comment:"Encrypted LeetCode password"`
}

type LeetCodeConfig struct {
	Site        Site        `yaml:"site" mapstructure:"site" comment:"LeetCode site, https://leetcode.com or https://leetcode.cn"`
	Credentials Credentials `yaml:"credential" mapstructure:"credential" comment:"Credential to access LeetCode"`
}

func (c *Config) ConfigDir() string {
	return c.dir
}

func (c *Config) GlobalConfigFile() string {
	return filepath.Join(c.dir, globalConfigFile)
}

func (c *Config) ProjectRoot() string {
	if c.projectRoot == "" {
		dir, _ := os.Getwd()
		c.projectRoot = dir
		for {
			if utils.IsExist(filepath.Join(dir, projectConfigFilename)) {
				c.projectRoot = dir
				break
			}
			dir1 := filepath.Dir(dir)
			// Reached root.
			if dir1 == dir {
				break
			}
			dir = dir1
		}
	}
	return c.projectRoot
}

func (c *Config) ProjectConfigFile() string {
	return filepath.Join(c.ProjectRoot(), projectConfigFilename)
}

func (c *Config) ProjectConfigFilename() string {
	return projectConfigFilename
}

func (c *Config) LeetCodeCacheFile() string {
	return filepath.Join(c.dir, leetcodeCacheFile)
}

func (c *Config) Write(w io.Writer) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	node, _ := toYamlNode(c)
	err := enc.Encode(node)
	return err
}

func Default() *Config {
	home, _ := homedir.Dir()
	author := "Bob"
	configDir := filepath.Join(home, ".config", CmdName)
	return &Config{
		dir:      configDir,
		Author:   author,
		Gen:      "go",
		Language: ZH,
		LeetCode: LeetCodeConfig{
			Site: LeetCodeCN,
			Credentials: Credentials{
				ReadFromBrowser: "chrome",
			},
		},
		Editor: Editor{
			Command: "vim",
			Args:    nil,
		},
		Cache: "json",
		Go: GoConfig{
			BaseLangConfig:   BaseLangConfig{OutDir: "go"},
			SeparatePackage:  true,
			FilenameTemplate: ``,
		},
		Python:     BaseLangConfig{OutDir: "python"},
		Cpp:        BaseLangConfig{OutDir: "cpp"},
		Java:       BaseLangConfig{OutDir: "java"},
		Rust:       BaseLangConfig{OutDir: "rust"},
		C:          BaseLangConfig{OutDir: "c"},
		CSharp:     BaseLangConfig{OutDir: "csharp"},
		JavaScript: BaseLangConfig{OutDir: "javascript"},
		Ruby:       BaseLangConfig{OutDir: "ruby"},
		Swift:      BaseLangConfig{OutDir: "swift"},
		Kotlin:     BaseLangConfig{OutDir: "kotlin"},
		PHP:        BaseLangConfig{OutDir: "php"},
		// Add more languages here
	}
}

func Get() *Config {
	if cfg == nil {
		return Default()
	}
	return cfg
}

func Set(c Config) {
	cfg = &c
}

func Verify(c *Config) error {
	if c.LeetCode.Site != LeetCodeCN && c.LeetCode.Site != LeetCodeUS {
		return fmt.Errorf("invalid site: %s", c.LeetCode.Site)
	}
	if c.Language != ZH && c.Language != EN {
		return fmt.Errorf("invalid language: %s", c.Language)
	}
	if c.Gen == "" {
		return fmt.Errorf("gen is empty")
	}
	if c.Cache != "json" && c.Cache != "sqlite" {
		return fmt.Errorf("invalid cache type: %s", c.Cache)
	}
	return nil
}
