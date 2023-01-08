# Leetgo

[![Go Report Card](https://goreportcard.com/badge/github.com/j178/leetgo)](https://goreportcard.com/report/github.com/j178/leetgo)
[![CI](https://github.com/j178/leetgo/actions/workflows/ci.yaml/badge.svg)](https://github.com/j178/leetgo/actions/workflows/ci.yaml)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://makeapullrequest.com)
[![Twitter Follow](https://img.shields.io/twitter/follow/niceoe)](https://twitter.com/niceoe)

中文 | [English](./README.md)

`leetgo` is a command line tool that generates skeleton code for LeetCode questions in many languages. You can run and debug test cases locally with your favorite IDE.
Then you can submit your code to LeetCode directly.

[![asciicast](https://asciinema.org/a/0sUG7psmMfgWqzy9rr57hrcnX.svg)](https://asciinema.org/a/0sUG7psmMfgWqzy9rr57hrcnX)

`leetgo` supports generating contest questions as well.

TODO: add a https://asciinema.org/

**This project is in its early development stage, and anything is likely to change.**

## Highlight of features

- Search for and view a question by its ID or slug.
- Generate skeleton code and testing code for a question.
- Run test cases on your local machine.
- Generate contest questions just in time.

## Language support

Currently, `leetgo` supports generating code and local test for the following languages:
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

### Install via `go install`
 
```shell
go install github.com/j178/leetgo@latest
```

### Install via `brew install`

```shell
brew install j178/tap/leetgo
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
  help                    Help about any command

Flags:
  -v, --version       version for leetgo
  -l, --lang string   language of code to generate: cpp, go, python ...
  -y, --yes           answer yes to all prompts
  -h, --help          help for leetgo

Use "leetgo [command] --help" for more information about a command.
```
<!-- END USAGE -->

## LeetCode Support

Leetgo uses LeetCode's GraphQL API to get questions and submit solutions. You need to provide your LeetCode session ID to authenticate.

Currently only `leetcode.cn` is supported. `leetcode.com` is under development.

## Configuration

Leetgo reads global configuration from `~/.config/leetgo/config.yaml` and local configuration from `leetgo.yaml` in your project root, which are generated automatically when you run `leetgo init`.
You can tweak the configuration to your liking.

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
  filename_template: '{{ .Id | padWithZero 4 }}{{ if .SlugIsMeaningful }}.{{ .Slug }}{{ end }}'
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
  python3:
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
    # How to provide credentials: browser, cookies, password or none
    from: browser
    # LeetCode cookie: LEETCODE_SESSION
    session: ""
    # LeetCode cookie: csrftoken
    csrftoken: ""
    # LeetCode username
    username: ""
    # Encrypted LeetCode password
    password: ""
contest:
  # Base dir to put generated contest questions
  out_dir: contest
# The editor to open generated files
editor:
  # Use a predefined editor: vim, vscode, goland
  # Set to 'none' to disable, set to 'custom' to provide your own command
  use: none
  # Custom command to open files
  command: ""
  # Arguments to the command
  args: []
```
<!-- END CONFIG -->

## Troubleshooting

If you encounter any problems, please run your command with `DEBUG` environment variable set to `1`, copy the command output and open an issue.

## Contributions welcome!

[Good first issues](https://github.com/j178/leetgo/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) are a great place to start.


## Credits

- https://github.com/EndlessCheng/codeforces-go
- https://github.com/clearloop/leetcode-cli
- https://github.com/budougumi0617/leetgode
