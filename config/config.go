package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/j178/leetgo/constants"

	"github.com/j178/leetgo/utils"
)

var (
	globalCfg *Config
	Debug     = os.Getenv("DEBUG") == "1"
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
	OutDir           string `yaml:"out_dir" mapstructure:"out_dir" comment:"Base dir to put generated contest questions"`
	FilenameTemplate string `yaml:"filename_template" mapstructure:"filename_template" comment:"Template to generate filename of the question"`
	OpenInBrowser    bool   `yaml:"open_in_browser" mapstructure:"open_in_browser" comment:"Open the contest page in browser after generating"`
}

type Editor struct {
	Use     string   `yaml:"use" mapstructure:"use" comment:"Use a predefined editor: vim, vscode, goland\nSet to 'none' to disable, set to 'custom' to provide your own command"`
	Command string   `yaml:"command" mapstructure:"command" comment:"Custom command to open files"`
	Args    []string `yaml:"args" mapstructure:"args" comment:"Arguments to the command"`
}

type Block struct {
	Name     string `yaml:"name" mapstructure:"name"`
	Template string `yaml:"template" mapstructure:"template"`
}

type Modifier struct {
	Name   string `yaml:"name" mapstructure:"name"`
	Script string `yaml:"script,omitempty" mapstructure:"script"`
}

type CodeConfig struct {
	Lang                    string         `yaml:"lang" mapstructure:"lang" comment:"Language of code generated for questions: go, python, ... \n(will be override by project config and flag --lang)"`
	FilenameTemplate        string         `yaml:"filename_template" mapstructure:"filename_template" comment:"The default template to generate filename (without extension), e.g. {{.Id}}.{{.Slug}}\nAvailable attributes: Id, Slug, Title, Difficulty, Lang, SlugIsMeaningful\nAvailable functions: lower, upper, trim, padWithZero, toUnderscore"`
	SeparateDescriptionFile bool           `yaml:"separate_description_file" mapstructure:"separate_description_file" comment:"Generate question description into a separate file"`
	Blocks                  []Block        `yaml:"blocks,omitempty" mapstructure:"blocks" comment:"Replace some blocks of the generated code"`
	Modifiers               []Modifier     `yaml:"modifiers,omitempty" mapstructure:"modifiers" comment:"Functions that modify the generated code"`
	Go                      GoConfig       `yaml:"go" mapstructure:"go"`
	Python                  PythonConfig   `yaml:"python3" mapstructure:"python3"`
	Cpp                     CppConfig      `yaml:"cpp" mapstructure:"cpp"`
	Rust                    RustConfig     `yaml:"rust" mapstructure:"rust"`
	Java                    BaseLangConfig `yaml:"java" mapstructure:"java"`
	// Add more languages here
}

type BaseLangConfig struct {
	OutDir                  string     `yaml:"out_dir" mapstructure:"out_dir"`
	FilenameTemplate        string     `yaml:"filename_template" mapstructure:"filename_template" comment:"Overrides the default code.filename_template"`
	SeparateDescriptionFile bool       `yaml:"separate_description_file,omitempty" mapstructure:"separate_description_file" comment:"Generate question description into a separate file"`
	Blocks                  []Block    `yaml:"blocks,omitempty" mapstructure:"blocks" comment:"Replace some blocks of the generated code"`
	Modifiers               []Modifier `yaml:"modifiers,omitempty" mapstructure:"modifiers" comment:"Functions that modify the generated code"`
}

type GoConfig struct {
	BaseLangConfig `yaml:",inline" mapstructure:",squash"`
}

type PythonConfig struct {
	BaseLangConfig `yaml:",inline" mapstructure:",squash"`
	Executable     string `yaml:"executable" mapstructure:"executable" comment:"Python executable to run the generated code"`
}

type CppConfig struct {
	BaseLangConfig `yaml:",inline" mapstructure:",squash"`
	CXX            string   `yaml:"cxx" mapstructure:"cxx" comment:"C++ compiler"`
	CXXFLAGS       []string `yaml:"cxxflags" mapstructure:"cxxflags" comment:"C++ compiler flags (our Leetcode I/O library implementation requires C++17)"`
}

type RustConfig struct {
	BaseLangConfig `yaml:",inline" mapstructure:",squash"`
}

type Credentials struct {
	From      string `yaml:"from" mapstructure:"from" comment:"How to provide credentials: browser, cookies, password or none"`
	Session   string `yaml:"session" mapstructure:"session" comment:"LeetCode cookie: LEETCODE_SESSION"`
	CsrfToken string `yaml:"csrftoken" mapstructure:"csrftoken" comment:"LeetCode cookie: csrftoken"`
	Username  string `yaml:"username" mapstructure:"username" comment:"LeetCode username"`
	Password  string `yaml:"password" mapstructure:"password" comment:"Encrypted LeetCode password"`
}

type LeetCodeConfig struct {
	Site        LeetcodeSite `yaml:"site" mapstructure:"site" comment:"LeetCode site, https://leetcode.com or https://leetcode.cn"`
	Credentials Credentials  `yaml:"credentials" mapstructure:"credentials" comment:"Credentials to access LeetCode"`
}

func (c *Config) ConfigDir() string {
	if c.dir == "" {
		home, _ := homedir.Dir()
		c.dir = filepath.Join(home, ".config", constants.CmdName)
	}
	return c.dir
}

func (c *Config) CacheDir() string {
	return filepath.Join(c.ConfigDir(), "cache")
}

func (c *Config) TempDir() string {
	return filepath.Join(os.TempDir(), constants.CmdName)
}

func (c *Config) GlobalConfigFile() string {
	return filepath.Join(c.ConfigDir(), constants.GlobalConfigFilename)
}

func (c *Config) ProjectRoot() string {
	if c.projectRoot == "" {
		dir, _ := os.Getwd()
		c.projectRoot = dir
		for {
			if utils.IsExist(filepath.Join(dir, constants.ProjectConfigFilename)) {
				c.projectRoot = dir
				break
			}
			parent := filepath.Dir(dir)
			// Reached root.
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return c.projectRoot
}

func (c *Config) ProjectConfigFile() string {
	return filepath.Join(c.ProjectRoot(), constants.ProjectConfigFilename)
}

func (c *Config) StateFile() string {
	return filepath.Join(c.CacheDir(), constants.StateFilename)
}

func (c *Config) QuestionCacheFile(ext string) string {
	return filepath.Join(c.CacheDir(), constants.QuestionCacheBaseName+ext)
}

func (c *Config) Write(w io.Writer, withComments bool) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	var err error
	if withComments {
		node, _ := toYamlNode(c)
		err = enc.Encode(node)
	} else {
		err = enc.Encode(c)
	}

	return err
}

func Default() *Config {
	return &Config{
		Author:   "Bob",
		Language: ZH,
		Code: CodeConfig{
			Lang:                    "go",
			FilenameTemplate:        `{{ .Id | padWithZero 4 }}{{ if .SlugIsMeaningful }}.{{ .Slug }}{{ end }}`,
			SeparateDescriptionFile: false,
			Modifiers: []Modifier{
				{Name: "removeUselessComments"},
			},
			Go: GoConfig{
				BaseLangConfig: BaseLangConfig{
					OutDir: "go",
					Modifiers: []Modifier{
						{Name: "removeUselessComments"},
						{Name: "changeReceiverName"},
						{Name: "addNamedReturn"},
						{Name: "addMod"},
					},
				},
			},
			Cpp: CppConfig{
				BaseLangConfig: BaseLangConfig{OutDir: "cpp"},
				CXX:            "g++",
				CXXFLAGS:       []string{"-O2", "-std=c++17"},
			},
			Python: PythonConfig{
				BaseLangConfig: BaseLangConfig{OutDir: "python"},
				Executable:     constants.DefaultPython,
			},
			Java: BaseLangConfig{OutDir: "java"},
			Rust: RustConfig{BaseLangConfig: BaseLangConfig{OutDir: "rust"}},
			// Add more languages here
		},
		LeetCode: LeetCodeConfig{
			Site: LeetCodeCN,
			Credentials: Credentials{
				From: "browser",
			},
		},
		Editor: Editor{
			Use: "none",
		},
		Contest: ContestConfig{
			OutDir:           "contest",
			FilenameTemplate: `{{ .ContestShortSlug }}/{{ .Id }}{{ if .SlugIsMeaningful }}.{{ .Slug }}{{ end }}`,
			OpenInBrowser:    true,
		},
	}
}

func Get() *Config {
	if globalCfg == nil {
		return Default()
	}
	return globalCfg
}

func verify(c *Config) error {
	if c.Language != ZH && c.Language != EN {
		return fmt.Errorf("invalid `language` value: %s", c.Language)
	}
	if c.Code.Lang == "" {
		return fmt.Errorf("`code.lang` is empty")
	}
	switch strings.ToLower(string(c.LeetCode.Site)) {
	case "cn":
		c.LeetCode.Site = LeetCodeCN
	case "us":
		c.LeetCode.Site = LeetCodeUS
	}
	if c.LeetCode.Site != LeetCodeCN && c.LeetCode.Site != LeetCodeUS {
		return fmt.Errorf("invalid `leetcode.site` value: %s", c.LeetCode.Site)
	}
	credentialFrom := map[string]bool{
		"browser":  true,
		"cookies":  true,
		"password": true,
		"none":     true,
	}
	if !credentialFrom[c.LeetCode.Credentials.From] {
		return fmt.Errorf("invalid `leetcode.credentials.from` value: %s", c.LeetCode.Credentials.From)
	}

	if c.LeetCode.Credentials.From == "cookies" {
		if c.LeetCode.Credentials.Session == "" {
			return fmt.Errorf("`leetcode.credentials.session` is empty")
		}
		if c.LeetCode.Credentials.CsrfToken == "" {
			return fmt.Errorf("`leetcode.credentials.csrftoken` is empty")
		}
	}

	if c.LeetCode.Credentials.From == "password" {
		if c.LeetCode.Site == LeetCodeUS {
			return fmt.Errorf("username/password authentication is not supported for leetcode.com")
		}
		if c.LeetCode.Credentials.Username == "" {
			return fmt.Errorf("`leetcode.credentials.username` is empty")
		}
		if c.LeetCode.Credentials.Password == "" {
			return fmt.Errorf("`leetcode.credentials.password` is empty")
		}
		if !strings.HasPrefix(
			c.LeetCode.Credentials.Password,
			vaultHeader,
		) {
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
	}

	return nil
}

func Load(init bool) error {
	if globalCfg != nil {
		return nil
	}

	// load global configuration
	cfg := Get()

	viper.SetConfigFile(cfg.GlobalConfigFile())
	err := viper.ReadInConfig()
	if err != nil {
		if os.IsNotExist(err) {
			if !init {
				log.Warn(
					"global config file not found, have you ran `leetgo init`?",
					"file",
					cfg.GlobalConfigFile(),
				)
			}
			return nil
		}
		return fmt.Errorf("load config file %s failed: %w", cfg.GlobalConfigFile(), err)
	}

	// Don't read project config if we are running `init` command
	if !init {
		// load project configuration
		viper.SetConfigFile(cfg.ProjectConfigFile())
		err = viper.MergeInConfig()
		if err != nil {
			if os.IsNotExist(err) {
				log.Warn(
					fmt.Sprintf("%s not found, use global config only", constants.ProjectConfigFilename),
					"file",
					cfg.GlobalConfigFile(),
				)
			} else {
				return fmt.Errorf("load config file %s failed: %w", cfg.ProjectConfigFile(), err)
			}
		}
	}

	err = viper.Unmarshal(cfg)
	if err != nil {
		return fmt.Errorf("unmarshal config failed: %s", err)
	}

	if err = verify(cfg); err != nil {
		return fmt.Errorf("verify config failed: %s", err)
	}

	globalCfg = cfg
	return nil
}
