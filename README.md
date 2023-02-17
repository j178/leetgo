**[中文](./README_zh.md) | English**

# Leetgo

[![Go Report Card](https://goreportcard.com/badge/github.com/j178/leetgo)](https://goreportcard.com/report/github.com/j178/leetgo)
[![CI](https://github.com/j178/leetgo/actions/workflows/ci.yaml/badge.svg)](https://github.com/j178/leetgo/actions/workflows/ci.yaml)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://makeapullrequest.com)
[![Discord](https://img.shields.io/discord/1069106479744962582?label=discord&logo=discord)](https://discord.gg/bHsEwQQj9m)
[![Twitter Follow](https://img.shields.io/twitter/follow/niceoe)](https://twitter.com/niceoe)

`leetgo` is a command-line tool for LeetCode that provides almost all the functionality of LeetCode, 
allowing you to do all of your LeetCode exercises without leaving the terminal. 
It can automatically generate **skeleton code and test cases**, support **local testing and debugging**, 
and you can use any IDE you like to solve problems. 

And `leetgo` also supports real-time generation of **contest questions**, submitting all questions at once, so your submissions are always one step ahead!

[![asciicast](https://asciinema.org/a/7R2lnZj7T0hEuJ49SH2lZ04NG.svg)](https://asciinema.org/a/7R2lnZj7T0hEuJ49SH2lZ04NG)

## Highlight of features

- Generate description, skeleton code and testing code for a question
- Customize the code template for generated code, use modifiers to pre-process code
- Execute test cases on your local machine
- Wait and generate contest questions just in time, test and submit all at once
- Support for both leetcode.com and leetcode.cn
- Automatically read cookies from browser, no need to enter password
- Automatically open question files in your favourite editor

## Language support

`leetgo` supports code generation for most languages, and local testing for some languages.

Local testing means that you can run the test cases on your local machine, so you can use a debugger to debug your code.

Local testing requires more work to implement for each language, so not all languages are supported.

<!-- BEGIN MATRIX -->
|  | Generation | Local testing |
| --- | --- | --- |
| Go | :white_check_mark: | :white_check_mark: |
| Python | :white_check_mark: | Not yet |
| C++ | :white_check_mark: | Not yet |
| Rust | :white_check_mark: | Not yet |
| Java | :white_check_mark: | Not yet |
| JavaScript | :white_check_mark: | Not yet |
| PHP | :white_check_mark: | Not yet |
| C | :white_check_mark: | Not yet |
| C# | :white_check_mark: | Not yet |
| Ruby | :white_check_mark: | Not yet |
| Swift | :white_check_mark: | Not yet |
| Kotlin | :white_check_mark: | Not yet |
| Bash | :white_check_mark: | Not yet |
| MySQL | :white_check_mark: | Not yet |
| MSSQL | :white_check_mark: | Not yet |
| Oracle | :white_check_mark: | Not yet |
<!-- END MATRIX -->
and many other languages are planned. (Help wanted, contributions welcome!)

## Installation

You can download the latest binary from the [release page](https://github.com/j178/leetgo/releases).

### Install via go
 
```shell
git clone git@github.com:j178/leetgo.git
cd leetgo && go install
```

### Install via [HomeBrew](https://brew.sh/) on macOS/Linux

```shell
brew install j178/tap/leetgo
```

### Install via [Scoop](https://scoop.sh/) on Windows

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
  fix                     Use OpenAI GPT-3 API to fix your solution code (just for fun)
  edit                    Open solution in editor
  extract                 Extract solution code from generated file
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

### Question Identifier

Many `leetgo` commands rely on `qid` to find the leetcode question. `qid` is a simplified question 
identifier defined by leetgo, which includes the following forms (using the two-sum problem as an example):

```shell
leetgo pick two-sum          # `two-sum` is the question slug
leetgo pick 1                # `1` is the question id
leetgo pick today            # `today` means daily question
leetgo contest weekly100     # `weekly100` means the 100th weekly contest
leetgo test last             # `last` means the last generated question
leetgo test weekly100/1      # `weekly100/1` means the first question of the 100th weekly contest
leetgo submit b100/2         # `b100/2` means the second question of the 100th biweekly contest
leetgo submit w99/           # `w99/` means all questions of the 99th biweekly contest (must keep the trailing slash)
leetgo test last/1           # `last/1` means the first question of the last generated contest
leetgo test last/            # `last/` means all questions of the last generated contest (must keep the trailing slash)
```

## Configuration

Leetgo uses two levels of configuration files, the global configuration file located at `~/.config/leetgo/config.yaml` and the local configuration file located at `leetgo.yaml` in the project root. 

These configuration files are created during the `leetgo init` process. 
The local configuration file in the project overrides the global configuration. 

It is generally recommended to use the global configuration as the default configuration and customize it in the project by modifying the `leetgo.yaml` file.

Here is the demonstration of complete configurations:

<details>
<summary>Click to expand</summary>

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
  # Functions that modify the generated code
  modifiers:
    - name: removeUselessComments
  go:
    out_dir: go
    # Overrides the default code.filename_template
    filename_template: ""
    # Replace some blocks of the generated code
    blocks:
      - name: beforeMarker
        template: |
          package main

          {{ if .NeedsDefinition -}} import . "github.com/j178/leetgo/testutils/go" {{- end }}
    # Functions that modify the generated code
    modifiers:
      - name: removeUselessComments
      - name: changeReceiverName
      - name: addNamedReturn
      - name: addMod
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
  # Template to generate filename of the question
  filename_template: '{{ .ContestShortSlug }}/{{ .Id }}{{ if .SlugIsMeaningful }}.{{ .Slug }}{{ end }}'
  # Open the contest page in browser after generating
  open_in_browser: true
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
</details>

## LeetCode Support

`leetgo` uses LeetCode's GraphQL API to retrieve questions and submit solutions. `leetgo` needs your LeetCode cookies to access the authenticated API.

There are three ways to make cookies available to `leetgo`:

- Read cookies from browser automatically.
  
  Currently, `leetgo` supports Chrome, FireFox, Safari[^1], Edge.
  If you want to support other browsers, please let us know!

  ```yaml
  leetcode:
    credentials:
      from: browser
  ```

- Provide cookies in config file.
  
  You can get your cookies named `LEETCODE_SESSION` and `csrftoken` from browser's developer tools.

  ```yaml
  leetcode:
    credentials:
      from: cookies
      session: xxx
      csrftoken: xx
  ```

- Provide username and password in config file.

  You need to run `leetgo config encrypt` to encrypt your password first, plain text password is **not allowed**.

  ```yaml
  leetcode:
    credentials:
      from: password
      username: xxx
      password: |
        $LEETGO_VAULT;1.1;AES256
        61393232326161303064373437376538646432623336363563623935333863653666623633376466
        3836633339643934383061363239333833333634373137620a303466626335633332393336326564
        31633231333934323165376362646630643132626130626136326163333133663762356264353564
        6562653462396335300a313761363531363961656364366634666562663061633161366463393339
        3963
  ```

> **Note**
> Password authentication is not recommended, and it is not supported by `leetcode.com`.

## Advanced Usage

### Templates

Several fields in leetgo's config file support templating. These fields are often suffixed with `_template`.
You can use custom template to generate your own filename, code, etc.

### Blocks

A code file is composed of different blocks, you can overwrite some of them to provide your own snippets.

| Available blocks |
| -- | 
| header |
| description |
| title |
| beforeMarker |
| beforeCode |
| code |
| afterCode |
| afterMarker |

For example:
```yaml
code:
lang: cpp
cpp:
  blocks:
  - name: beforeCode
    template: |
      #include <iostream>
      using namespace std;
  - name: afterMarker
    template: |
      int main() {}
```

### Scripting

`leetgo` supports providing a JavaScript function to handle the code before generation, for example:

```yaml
code:
  lang: cpp
  cpp:
    modifiers:
    - name: removeUselessComments
    - script: |
        function modify(code) {
          return "// hello world\n" + code;
        } 
```

## FAQ

If you encounter any problems, please run your command with the `DEBUG` environment variable set to `1`, copy the command output, and open an issue.

Some common problems can be found in the [Q&A](https://github.com/j178/leetgo/discussions/categories/q-a) page.

## Contributions welcome!

[Good First Issues](https://github.com/j178/leetgo/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) are a good place to start, 
and you can also check out some [Help Wanted](https://github.com/j178/leetgo/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22) issues.

Before submitting a PR, please run `golangci-lint run --fix` to fix lint errors.

## Credits

Here are some awesome projects that inspired me to create this project:

- https://github.com/EndlessCheng/codeforces-go
- https://github.com/clearloop/leetcode-cli
- https://github.com/budougumi0617/leetgode
- https://github.com/skygragon/leetcode-cli

[^1]: For Safari on MacOS, you may need to grant `Full Disk Access` privilege to your terminal app which you would like to run `leetgo`.
