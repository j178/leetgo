package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/j178/leetgo/utils"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

const (
	CmdName               = "leetgo"
	globalConfigFile      = "config.yaml"
	projectConfigFilename = CmdName + ".yaml"
	leetcodeCacheFile     = "cache/leetcode-questions.db"
	stateFile             = "cache/state.json"
	codeBeginMark         = "@lc code=start"
	codeEndMark           = "@lc code=end"
)

var (
	cfg   *Config
	Debug = os.Getenv("DEBUG") != ""
)

type (
	LeetcodeSite string
	Language     string
)

const (
	LeetCodeCN LeetcodeSite = "https://leetcode.cn"
	LeetCodeUS LeetcodeSite = "https://leetcode.com"
	ZH         Language     = "zh"
	EN         Language     = "en"
)

type Config struct {
	dir         string
	projectRoot string
	Author      string         `yaml:"author" mapstructure:"author" comment:"Your name"`
	Language    Language       `yaml:"language" mapstructure:"language" comment:"Language of the question description: zh or en"`
	Code        CodeConfig     `yaml:"code" mapstructure:"code" comment:"Code configuration"`
	LeetCode    LeetCodeConfig `yaml:"leetcode" mapstructure:"leetcode" comment:"LeetCode configuration"`
	Contest     ContestConfig  `yaml:"contest" mapstructure:"contest"`
	Editor      Editor         `yaml:"editor" mapstructure:"editor" comment:"The editor to open generated files"`
}

type ContestConfig struct {
	OutDir string `yaml:"out_dir" mapstructure:"out_dir" comment:"Base dir to put generated contest questions"`
}

type Editor struct {
	Use     string   `yaml:"use" mapstructure:"use" comment:"Use a predefined editor: vim, vscode, goland, set to none to disable opening files after generation"`
	Command string   `yaml:"command" mapstructure:"command" comment:"Custom command to open files"`
	Args    []string `yaml:"args" mapstructure:"args" comment:"Arguments to the command"`
}

type CodeConfig struct {
	Lang             string         `yaml:"lang" mapstructure:"lang" comment:"Language of code generated for questions: go, python, ... \n(will be override by project config and flag --lang)"`
	FilenameTemplate string         `yaml:"filename_template" mapstructure:"filename_template" comment:"The default template to generate filename (without extension), e.g. {{.Id}}.{{.Slug}}\nAvailable attributes: Id, Slug, Title, Difficulty, Lang, SlugIsMeaningful\nAvailable functions: lower, upper, trim, padWithZero, toUnderscore"`
	CodeBeginMark    string         `yaml:"code_begin_mark" mapstructure:"code_begin_mark" comment:"The mark to indicate the beginning of the code"`
	CodeEndMark      string         `yaml:"code_end_mark" mapstructure:"code_end_mark" comment:"The mark to indicate the end of the code"`
	Go               GoConfig       `yaml:"go" mapstructure:"go"`
	Python           BaseLangConfig `yaml:"python" mapstructure:"python"`
	Cpp              BaseLangConfig `yaml:"cpp" mapstructure:"cpp"`
	Java             BaseLangConfig `yaml:"java" mapstructure:"java"`
	Rust             BaseLangConfig `yaml:"rust" mapstructure:"rust"`
	// Add more languages here
}

type BaseLangConfig struct {
	OutDir           string `yaml:"out_dir" mapstructure:"out_dir"`
	FilenameTemplate string `yaml:"filename_template" mapstructure:"filename_template" comment:"Overrides the default code.filename_template"`
}

type GoConfig struct {
	BaseLangConfig `yaml:",inline" mapstructure:",squash"`
	GoModPath      string `yaml:"go_mod_path" mapstructure:"go_mod_path" comment:"Go module path for the generated code"`
}

type Credentials struct {
	ReadFromBrowser string `yaml:"read_from_browser" mapstructure:"read_from_browser" comment:"Read leetcode cookie from browser, currently only chrome is supported."`
	Session         string `yaml:"session,omitempty" mapstructure:"session" comment:"LeetCode cookie: LEETCODE_SESSION"`
	CsrfToken       string `yaml:"csrf_token,omitempty" mapstructure:"csrf_token" comment:"LeetCode cookie: csrftoken"`
	Username        string `yaml:"username,omitempty" mapstructure:"username" comment:"LeetCode username"`
	Password        string `yaml:"password,omitempty" mapstructure:"password" comment:"Encrypted LeetCode password"`
}

type LeetCodeConfig struct {
	Site        LeetcodeSite `yaml:"site" mapstructure:"site" comment:"LeetCode site, https://leetcode.com or https://leetcode.cn"`
	Credentials Credentials  `yaml:"credentials" mapstructure:"credentials" comment:"Credentials to access LeetCode"`
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

func (c *Config) StateFile() string {
	return filepath.Join(c.dir, stateFile)
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
		Language: ZH,
		Code: CodeConfig{
			Lang:             "go",
			CodeBeginMark:    codeBeginMark,
			CodeEndMark:      codeEndMark,
			FilenameTemplate: `{{ .Id | padWithZero 4 }}{{ if .SlugIsMeaningful }}.{{ .Slug | toUnderscore }}{{ end }}`,
			Go: GoConfig{
				BaseLangConfig: BaseLangConfig{OutDir: "go"},
			},
			Python: BaseLangConfig{OutDir: "python"},
			Cpp:    BaseLangConfig{OutDir: "cpp"},
			Java:   BaseLangConfig{OutDir: "java"},
			Rust:   BaseLangConfig{OutDir: "rust"},
			// Add more languages here
		},
		LeetCode: LeetCodeConfig{
			Site: LeetCodeCN,
			Credentials: Credentials{
				ReadFromBrowser: "chrome",
			},
		},
		Editor: Editor{
			Use: "none",
		},
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
	if c.LeetCode.Site == LeetCodeUS && c.LeetCode.Credentials.Password != "" {
		return fmt.Errorf("username/password authentication is not supported for leetcode.com")
	}
	if c.LeetCode.Credentials.Password != "" && !strings.HasPrefix(c.LeetCode.Credentials.Password, vaultHeader) {
		return fmt.Errorf("password is not encrypted, you need to run `leetgo config encrypt` before put it in config file")
	}
	pw := c.LeetCode.Credentials.Password
	if pw != "" {
		var err error
		c.LeetCode.Credentials.Password, err = Decrypt(pw)
		if err != nil {
			return err
		}
	}
	if c.LeetCode.Credentials.ReadFromBrowser != "chrome" {
		return fmt.Errorf("invalid leetcode.credentials.read_from_browser: %s", c.LeetCode.Credentials.ReadFromBrowser)
	}
	if c.Language != ZH && c.Language != EN {
		return fmt.Errorf("invalid language: %s", c.Language)
	}
	if c.Code.Lang == "" {
		return fmt.Errorf("code.lang is empty")
	}
	return nil
}
