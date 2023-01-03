# Leetgo

[![Go Report Card](https://goreportcard.com/badge/github.com/j178/leetgo)](https://goreportcard.com/report/github.com/j178/leetgo)
[![CI](https://github.com/j178/leetgo/actions/workflows/ci.yaml/badge.svg)](https://github.com/j178/leetgo/actions/workflows/ci.yaml)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://makeapullrequest.com)
[![Twitter Follow](https://img.shields.io/twitter/follow/niceoe)](https://twitter.com/niceoe)

*This project is still in its early development stage and many features have not yet been implemented. Everything is subject to rapid change.*

[中文](./README_zh.md) | English

`leetgo` is a command-line tool for LeetCode that provides almost all the functionality of LeetCode, 
allowing you to complete all of your LeetCode exercises without leaving the terminal. 
It can automatically generate **skeleton code and test cases**, support **local testing and debugging**, 
and you can use any IDE you like to solve problems. 

And `leetgo` also supports real-time generation of **contest questions**, so your submissions are one step ahead!

[![asciicast](https://asciinema.org/a/0sUG7psmMfgWqzy9rr57hrcnX.svg)](https://asciinema.org/a/0sUG7psmMfgWqzy9rr57hrcnX)


## Highlight of features

- Pick question with a simple and fancy UI.
- Generate skeleton code and testing code for a question.
- Run test cases on your local machine.
- Generate contest questions just in time.
- Support both leetcode.com and leetcode.cn
- Read cookies from browser automatically, no need to provide password.

## Language support

`leetgo` supports generating code for most languages, and local test skeleton for some languages.

Local test means that you can run the test cases on your local machine, so you can use a debugger to debug your code.

Local test requires more work to implement for each language, so not all languages are supported.

<!-- BEGIN MATRIX -->
|  | Generate | Local Test |
| --- | --- | --- |
| Go | :white_check_mark: | :white_check_mark: |
| Python | :white_check_mark: | :x: |
| C++ | :white_check_mark: | :x: |
| Rust | :white_check_mark: | :x: |
| Java | :white_check_mark: | :x: |
| JavaScript | :white_check_mark: | :x: |
| PHP | :white_check_mark: | :x: |
| C | :white_check_mark: | :x: |
| C# | :white_check_mark: | :x: |
| Ruby | :white_check_mark: | :x: |
| Swift | :white_check_mark: | :x: |
| Kotlin | :white_check_mark: | :x: |
<!-- END MATRIX -->
and many other languages are in plan. (help wanted, contributions welcome!)

## Installation

You can download the latest binary from the [release page](https://github.com/j178/leetgo/releases).

### Install via `go`
 
```shell
go install github.com/j178/leetgo@latest
```

### Install via `brew`

```shell
brew install j178/tap/leetgo
```

### Install via `scoop`

```shell
scoop bucket add j178 https://github.com/j178/scoop-bucket.git
scoop install j178/leetgo
```

## Usage
<!-- BEGIN USAGE -->
```
Usage:
  leetgo [command]

Available Commands:
  init                    Init a leetcode workspace
  pick                    Generate a new question
  info                    Show question info
  test                    Run question test cases
  submit                  Submit solution
  contest                 Generate contest questions
  cache                   Manage local questions cache
  config                  Show configurations
  encrypt                 Encrypt a sensitive string to be used in config file
  help                    Help about any command

Flags:
  -v, --version       version for leetgo
  -l, --lang string   language of code to generate: cpp, go, python ...
  -y, --yes           answer yes to all prompts
  -h, --help          help for leetgo

Use "leetgo [command] --help" for more information about a command.
```
<!-- END USAGE -->

## Configuration

Leetgo uses two levels of configuration files, the global configuration file located at `~/.config/leetgo/config.yaml` and the local configuration file located at `leetgo.yaml` in the project root. 

These configuration files are generated during the `leetgo init` process. 
The local configuration file in the project will override the global configuration. 

It is generally recommended to use the global configuration as the default configuration and customize it in the project by modifying the `leetgo.yaml` file.

Here is the demonstration of full configurations:

<!-- BEGIN CONFIG -->
```yaml
# Your name
author: Bob
# Language of the question description: zh or en
language: zh
# Code configuration
code:
  # Language of code generated for questions: go, python, ... 
  # (will be override by project config and flag --lang)
  lang: go
  # The default template to generate filename (without extension), e.g. {{.Id}}.{{.Slug}}
  # Available attributes: Id, Slug, Title, Difficulty, Lang, SlugIsMeaningful
  # Available functions: lower, upper, trim, padWithZero, toUnderscore
  filename_template: '{{ .Id | padWithZero 4 }}{{ if .SlugIsMeaningful }}.{{ .Slug | toUnderscore }}{{ end }}'
  # The mark to indicate the beginning of the code
  code_begin_mark: '@lc code=start'
  # The mark to indicate the end of the code
  code_end_mark: '@lc code=end'
  go:
    out_dir: go
    # Overrides the default code.filename_template
    filename_template: ""
    # Go module path for the generated code
    go_mod_path: ""
  python:
    out_dir: python
    # Overrides the default code.filename_template
    filename_template: ""
  cpp:
    out_dir: cpp
    # Overrides the default code.filename_template
    filename_template: ""
  java:
    out_dir: java
    # Overrides the default code.filename_template
    filename_template: ""
  rust:
    out_dir: rust
    # Overrides the default code.filename_template
    filename_template: ""
# LeetCode configuration
leetcode:
  # LeetCode site, https://leetcode.com or https://leetcode.cn
  site: https://leetcode.cn
  # Credentials to access LeetCode
  credentials:
    # Read leetcode cookie from browser, currently only chrome is supported.
    read_from_browser: chrome
contest:
  # Base dir to put generated contest questions
  out_dir: ""
# The editor to open generated files
editor:
  # Use a predefined editor: vim, vscode, goland, set to none to disable opening files after generation
  use: none
  # Custom command to open files
  command: ""
  # Arguments to the command
  args: []
# Cache type, json or sqlite
cache: json
```
<!-- END CONFIG -->

## LeetCode Support

Currently only `leetcode.cn` is supported. Support for `leetcode.com` is under development.

`leetgo` uses LeetCode's GraphQL API to get questions and submit solutions. `leetgo` needs your LeetCode cookies to access authenticated API.

There are three ways to provide cookies to `leetgo`:

- Read cookies from browser automatically.
  
  Currently only chrome is supported, if you want to support other browsers, please let us know!

  ```yaml
  leetcode:
    credentials:
      read_from_browser: chrome
  ```

- Provide cookies in config file.
  
  You can get your cookies named `LEETCODE_SESSION` and `csrftoken` from browser's developer tools.

  ```yaml
  leetcode:
    credentials:
      session: xxx
      csrftoken: xx
  ```

- Provide username and password in config file.

  You need to run `leetgo config encrypt` to encrypt your password first, plain text password is **not allowed**.

  ```yaml
  leetcode:
    credentials:
      username: xxx
      password: |
        $LEETGO_VAULT;1.1;AES256
        61393232326161303064373437376538646432623336363563623935333863653666623633376466
        3836633339643934383061363239333833333634373137620a303466626335633332393336326564
        31633231333934323165376362646630643132626130626136326163333133663762356264353564
        6562653462396335300a313761363531363961656364366634666562663061633161366463393339
        3963
  ```

**Note**: username/password authentication is not recommended, and it is not supported by `leetcode.com`.

## Troubleshooting

If you encounter any problems, please run your command with `DEBUG` environment variable set to `1`, copy the command output and open an issue.

## Contributions welcome!

[Good first issues](https://github.com/j178/leetgo/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) are a great place to start, 
and you can also check out some [help wanted](https://github.com/j178/leetgo/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22) issues.

## Credits

Here are some awesome projects that inspired me to create this project:

- https://github.com/EndlessCheng/codeforces-go
- https://github.com/clearloop/leetcode-cli
- https://github.com/budougumi0617/leetgode
- https://github.com/skygragon/leetcode-cli
