package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/shlex"
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
	Code        CodeConfig     `yaml:"code" mapstructure:"code"`
	LeetCode    LeetCodeConfig `yaml:"leetcode" mapstructure:"leetcode"`
	Contest     ContestConfig  `yaml:"contest" mapstructure:"contest"`
	Editor      Editor         `yaml:"editor" mapstructure:"editor" comment:"Editor settings to open generated files"`
}

type ContestConfig struct {
	OutDir           string `yaml:"out_dir" mapstructure:"out_dir" comment:"Base dir to put generated contest questions"`
	FilenameTemplate string `yaml:"filename_template" mapstructure:"filename_template" comment:"Template to generate filename of the question"`
	OpenInBrowser    bool   `yaml:"open_in_browser" mapstructure:"open_in_browser" comment:"Open the contest page in browser after generating"`
}

type Editor struct {
	Use     string `yaml:"use" mapstructure:"use" comment:"Use a predefined editor: vim, vscode, goland\nSet to 'none' to disable, set to 'custom' to provide your own command"`
	Command string `yaml:"command" mapstructure:"command" comment:"Custom command to open files"`
	Args    string `yaml:"args" mapstructure:"args" comment:"Arguments to the command.\nString contains {{.CodeFile}}, {{.TestFile}}, {{.DescriptionFile}}, {{.TestCasesFile}} will be replaced with corresponding file path.\n{{.Folder}} will be substituted with the output directory.\n{{.Files}} will be substituted with the list of all file paths."`
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
	Lang                    string       `yaml:"lang" mapstructure:"lang" comment:"Language of code generated for questions: cpp, python, go, java... \n(will be overridden by command line flag -l/--lang)"`
	FilenameTemplate        string       `yaml:"filename_template" mapstructure:"filename_template" comment:"The default template to generate filename (without extension), e.g. {{.Id}}.{{.Slug}}\nAvailable attributes: Id, Slug, Title, Difficulty, Lang, SlugIsMeaningful\nAvailable functions: lower, upper, trim, padWithZero, toUnderscore, group"`
	SeparateDescriptionFile bool         `yaml:"separate_description_file" mapstructure:"separate_description_file" comment:"Generate question description into a separate question.md file"`
	Blocks                  []Block      `yaml:"blocks,omitempty" mapstructure:"blocks" comment:"Default block definitions for all languages"`
	Modifiers               []Modifier   `yaml:"modifiers,omitempty" mapstructure:"modifiers" comment:"Default modifiers for all languages"`
	Go                      GoConfig     `yaml:"go" mapstructure:"go"`
	Python                  PythonConfig `yaml:"python3" mapstructure:"python3"`
	Cpp                     CppConfig    `yaml:"cpp" mapstructure:"cpp"`
	Rust                    RustConfig   `yaml:"rust" mapstructure:"rust"`
	Java                    JavaConfig   `yaml:"java" mapstructure:"java"`
	// Add more languages here
}

type BaseLangConfig struct {
	OutDir                  string     `yaml:"out_dir" mapstructure:"out_dir"`
	FilenameTemplate        string     `yaml:"filename_template,omitempty" mapstructure:"filename_template" comment:"Overrides the default code.filename_template, empty will be ignored"`
	SeparateDescriptionFile bool       `yaml:"separate_description_file,omitempty" mapstructure:"separate_description_file" comment:"Generate question description into a separate question.md file"`
	Blocks                  []Block    `yaml:"blocks,omitempty" mapstructure:"blocks" comment:"Replace some blocks of the generated code"`
	Modifiers               []Modifier `yaml:"modifiers,omitempty" mapstructure:"modifiers" comment:"Functions that modify the generated code"`
}

type GoConfig struct {
	BaseLangConfig `yaml:",inline" mapstructure:",squash"`
}

type PythonConfig struct {
	BaseLangConfig `yaml:",inline" mapstructure:",squash"`
	Executable     string `yaml:"executable" mapstructure:"executable" comment:"Python executable that creates the venv"`
}

type CppConfig struct {
	BaseLangConfig `yaml:",inline" mapstructure:",squash"`
	CXX            string `yaml:"cxx" mapstructure:"cxx" comment:"C++ compiler"`
	CXXFLAGS       string `yaml:"cxxflags" mapstructure:"cxxflags" comment:"C++ compiler flags (our Leetcode I/O library implementation requires C++17)"`
}

type RustConfig struct {
	BaseLangConfig `yaml:",inline" mapstructure:",squash"`
}

type JavaConfig struct {
	BaseLangConfig `yaml:",inline" mapstructure:",squash"`
	GroupID        string `yaml:"group_id" mapstructure:"group_id" comment:"Maven group ID"`
}

type Credentials struct {
	From     string   `yaml:"from" mapstructure:"from" comment:"How to provide credentials: browser, cookies, password or none"`
	Browsers []string `yaml:"browsers" mapstructure:"browsers" comment:"Browsers to get cookies from: chrome, safari, edge or firefox. If empty, all browsers will be tried"`
}

type LeetCodeConfig struct {
	Site        LeetcodeSite `yaml:"site" mapstructure:"site" comment:"LeetCode site, https://leetcode.com or https://leetcode.cn"`
	Credentials Credentials  `yaml:"credentials" mapstructure:"credentials" comment:"Credentials to access LeetCode"`
}

func (c *Config) HomeDir() string {
	if c.dir != "" {
		return c.dir
	}
	if home := os.Getenv("LEETGO_HOME"); home != "" {
		c.dir, _ = filepath.Abs(home)
	}
	if c.dir == "" {
		home, _ := homedir.Dir()
		c.dir = filepath.Join(home, ".config", constants.CmdName)
	}
	return c.dir
}

func (c *Config) CacheDir() string {
	return filepath.Join(c.HomeDir(), "cache")
}

func (c *Config) TempDir() string {
	return filepath.Join(os.TempDir(), constants.CmdName)
}

func (c *Config) ProjectRoot() string {
	if c.projectRoot == "" {
		dir, _ := os.Getwd()
		c.projectRoot = dir
		for {
			if utils.IsExist(filepath.Join(dir, constants.ConfigFilename)) {
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

func (c *Config) ConfigFile() string {
	return filepath.Join(c.ProjectRoot(), constants.ConfigFilename)
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

func defaultConfig() *Config {
	return &Config{
		Author:   "Bob",
		Language: ZH,
		Code: CodeConfig{
			Lang:                    "go",
			FilenameTemplate:        `{{ .Id | padWithZero 4 }}{{ if .SlugIsMeaningful }}.{{ .Slug }}{{ end }}`,
			SeparateDescriptionFile: true,
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
				CXXFLAGS:       "-O2 -std=c++17",
			},
			Python: PythonConfig{
				BaseLangConfig: BaseLangConfig{OutDir: "python"},
				Executable:     constants.DefaultPython,
			},
			Java: JavaConfig{
				BaseLangConfig: BaseLangConfig{
					OutDir:           "java",
					FilenameTemplate: `p{{ .Id | padWithZero 4 }}{{ if .SlugIsMeaningful }}_{{ .Slug | lower | toUnderscore }}{{ end }}`,
				},
			},
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
		return defaultConfig()
	}
	return globalCfg
}

// define here to avoid dependency on `leetcode` package.
var credentialFrom = map[string]bool{
	"browser":  true,
	"cookies":  true,
	"password": true,
	"none":     true,
}

func verify(c *Config) error {
	if c.Language != ZH && c.Language != EN {
		return fmt.Errorf("invalid `language` value: %s, only `en` or `zh` is supported", c.Language)
	}
	if c.Code.Lang == "" {
		return fmt.Errorf("`code.lang` not set, please set it in config file or provide it with `--lang/-l` flag")
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

	if !credentialFrom[c.LeetCode.Credentials.From] {
		return fmt.Errorf("invalid `leetcode.credentials.from` value: %s", c.LeetCode.Credentials.From)
	}
	if c.LeetCode.Credentials.From == "password" && c.LeetCode.Site == LeetCodeUS {
		return errors.New("username/password authentication is not supported for leetcode.com")
	}

	if c.Editor.Args != "" {
		if _, err := shlex.Split(c.Editor.Args); err != nil {
			return fmt.Errorf("invalid `editor.args`: %w", err)
		}
	}
	if c.Code.Cpp.CXXFLAGS != "" {
		if _, err := shlex.Split(c.Code.Cpp.CXXFLAGS); err != nil {
			return fmt.Errorf("invalid `code.cpp.cxxflags`: %w", err)
		}
	}
	return nil
}

func Load(init bool) error {
	if globalCfg != nil {
		return nil
	}

	// load default configuration
	cfg := defaultConfig()

	viper.SetConfigType("yaml")
	cfgBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal default config failed: %w", err)
	}
	err = viper.ReadConfig(bytes.NewBuffer(cfgBytes))
	if err != nil {
		return fmt.Errorf("read default config failed: %w", err)
	}

	// load project configuration
	if !init {
		viper.SetConfigFile(cfg.ConfigFile())
		err = viper.MergeInConfig()
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("%s not found, run `leetgo init` first", constants.ConfigFilename)
			}
			return fmt.Errorf("load config file %s failed: %w", cfg.ConfigFile(), err)
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
