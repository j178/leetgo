**[中文](./README_zh.md) | English**

# Leetgo

[![Go Report Card](https://goreportcard.com/badge/github.com/j178/leetgo)](https://goreportcard.com/report/github.com/j178/leetgo)
[![CI](https://github.com/j178/leetgo/actions/workflows/ci.yaml/badge.svg)](https://github.com/j178/leetgo/actions/workflows/ci.yaml)
[![GitHub downloads](https://img.shields.io/github/downloads/j178/leetgo/total)](https://github.com/j178/leetgo/releases)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://makeapullrequest.com)
[![Twitter Follow](https://img.shields.io/twitter/follow/niceoe)](https://twitter.com/niceoe)

`leetgo` is a command-line tool for LeetCode that provides almost all the functionality of LeetCode,
allowing you to do all of your LeetCode exercises without leaving the terminal.
It can automatically generate **skeleton code and test cases**, support **local testing and debugging**,
and you can use any IDE you like to solve problems.

And `leetgo` also supports real-time generation of **contest questions**, submitting all questions at once, so your submissions are always one step ahead!

## Quick Start

1. [Install leetgo](#installation)
2. Initialize leetgo: `leetgo init -t <us or cn> -l <lang>`
3. Edit leetgo config file: `leetgo.yaml` and `~/.config/leetgo/config.yaml`
4. Pick a question: `leetgo pick <id or name or today>`
5. Test your code: `leetgo test last -L`
6. Submit your code: `leetgo submit last`

You can test and submit in one command: `leetgo test last -L -s`

You can edit the question file in your favorite editor: `leetgo edit last`

## Demo

![demo](https://github.com/j178/leetgo/assets/10510431/8eaee981-a1f7-4b40-b9df-5af3c72daf26)

## Features

- Generate description, skeleton code and testing code for a question
- Customize the code template for generated code, use modifiers to pre-process code
- Execute test cases on your local machine
- Wait and generate contest questions just in time, test and submit all at once
- Support for both leetcode.com and leetcode.cn
- Automatically read cookies from browser, no need to enter password
- Automatically open question files in your favourite editor
- Use OpenAI to automatically discover and fix issues in the code (Experimental)

## Language support

`leetgo` supports code generation for most languages, and local testing for some languages.

In the Go language, running `leetgo pick 257` will generate the following code:

```go
// Omitted some code...
// @lc code=begin

func binaryTreePaths(root *TreeNode) (ans []string) {

	return
}

// @lc code=end

func main() {
	stdin := bufio.NewReader(os.Stdin)
	root := Deserialize[*TreeNode](ReadLine(stdin))
	ans := binaryTreePaths(root)
	fmt.Println("output: " + Serialize(ans))
}
```

This is a complete and runnable program. You can run it directly, input the test cases, and compare the results. 
`leetgo test -L` will automatically run this program with the test cases in `testcases.txt` and compare the results.

Local testing means that you can run the test cases on your local machine, so you can use a debugger to debug your code.

Local testing requires more work to implement for each language, so not all languages are supported. Below is the current support matrix:

<!-- BEGIN MATRIX -->
|  | Generation | Local testing |
| --- | --- | --- |
| Go | :white_check_mark: | :white_check_mark: |
| Python | :white_check_mark: | :white_check_mark: |
| C++ | :white_check_mark: | :white_check_mark: |
| Rust | :white_check_mark: | :white_check_mark: |
| Java | :white_check_mark: | Not yet |
| JavaScript | :white_check_mark: | Not yet |
| TypeScript | :white_check_mark: | Not yet |
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
| Erlang | :white_check_mark: | Not yet |
| Racket | :white_check_mark: | Not yet |
| Scala | :white_check_mark: | Not yet |
| Elixir | :white_check_mark: | Not yet |
| Dart | :white_check_mark: | Not yet |
<!-- END MATRIX -->

Welcome to help us implement local testing for more languages!

## Installation

You can download the latest binary from the [release page](https://github.com/j178/leetgo/releases).

### Install via [HomeBrew](https://brew.sh/) on macOS/Linux

```shell
brew install j178/tap/leetgo
```

### Install via [Scoop](https://scoop.sh/) on Windows

```shell
scoop bucket add j178 https://github.com/j178/scoop-bucket.git
scoop install j178/leetgo
```

### Install via go

```shell
go install github.com/j178/leetgo@latest
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
  fix                     Use ChatGPT API to fix your solution code (just for fun)
  edit                    Open solution in editor
  contest                 Generate contest questions
  cache                   Manage local questions cache
  debug                   Show debug info
  open                    Open one or multiple question pages in a browser
  help                    Help about any command

Flags:
  -v, --version       version for leetgo
  -l, --lang string   language of code to generate: cpp, go, python ...
      --site string   leetcode site: cn, us
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
leetgo pick yesterday        # `yesterday` means the question of yesterday
leetgo pick today-1          # `today-1` means the question of yesterday, same as `yesterday`. `today-2`, `today-3` etc are also supported.
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
code:
  # Language of code generated for questions: go, python, ... 
  # (will be override by project config and flag --lang)
  lang: go
  # The default template to generate filename (without extension), e.g. {{.Id}}.{{.Slug}}
  # Available attributes: Id, Slug, Title, Difficulty, Lang, SlugIsMeaningful
  # Available functions: lower, upper, trim, padWithZero, toUnderscore, group
  filename_template: '{{ .Id | padWithZero 4 }}{{ if .SlugIsMeaningful }}.{{ .Slug }}{{ end }}'
  # Default setting for separate_description_file
  separate_description_file: true
  # Default modifiers for all languages
  modifiers:
    - name: removeUselessComments
  go:
    out_dir: go
    # Functions that modify the generated code
    modifiers:
      - name: removeUselessComments
      - name: changeReceiverName
      - name: addNamedReturn
      - name: addMod
  python3:
    out_dir: python
    # Python executable that creates the venv
    executable: python3
  cpp:
    out_dir: cpp
    # C++ compiler
    cxx: g++
    # C++ compiler flags (our Leetcode I/O library implementation requires C++17)
    cxxflags: -O2 -std=c++17
  rust:
    out_dir: rust
  java:
    out_dir: java
leetcode:
  # LeetCode site, https://leetcode.com or https://leetcode.cn
  site: https://leetcode.cn
  # Credentials to access LeetCode
  credentials:
    # How to provide credentials: browser, cookies, password or none
    from: browser
    # Browsers to get cookies from: chrome, safari, edge or firefox. If empty, all browsers will be tried
    browsers: []
contest:
  # Base dir to put generated contest questions
  out_dir: contest
  # Template to generate filename of the question
  filename_template: '{{ .ContestShortSlug }}/{{ .Id }}{{ if .SlugIsMeaningful }}.{{ .Slug }}{{ end }}'
  # Open the contest page in browser after generating
  open_in_browser: true
# Editor settings to open generated files
editor:
  # Use a predefined editor: vim, vscode, goland
  # Set to 'none' to disable, set to 'custom' to provide your own command
  use: none
  # Custom command to open files
  command: ""
  # Arguments to the command.
  # String contains {{.CodeFile}}, {{.TestFile}}, {{.DescriptionFile}}, {{.TestCasesFile}} will be replaced with corresponding file path.
  # {{.Folder}} will be substituted with the output directory.
  # {{.Files}} will be substituted with the list of all file paths.
  args: ""
```
<!-- END CONFIG -->
</details>

## LeetCode Support

`leetgo` uses LeetCode's GraphQL API to retrieve questions and submit solutions. `leetgo` needs your LeetCode cookies to access the authenticated API.

There are three ways to make cookies available to `leetgo`:

- Read cookies from browser automatically.

  Currently, `leetgo` supports Chrome, FireFox, Safari[^1], Edge.

  ```yaml
  leetcode:
    credentials:
      from: browser
  ```

- Provide cookies.

  You can get your cookies named `LEETCODE_SESSION` and `csrftoken` from browser's developer tools, and set the `LEETCODE_SESSION` and `LEETCODE_CSRFTOKEN` environment variables.

  ```yaml
  leetcode:
    credentials:
      from: cookies
  ```

- Provide username and password through `LEETCODE_USERNAME` and `LEETCODE_PASSWORD` environment variables.

  ```yaml
  leetcode:
    credentials:
      from: password
  ```

> **Note**
> Password authentication is not recommended, and it is not supported by `leetcode.com`.

## Advanced Usage

### `testcases.txt`

`testcasts.txt` is generated by `leetgo` and contains all the test cases of the question.

You can add a new test case by specifying only the input and leaving the output blank. When you run `leetgo test` (without `-L`), the expected output will be retrieved from the remote server.
For example:

```
input:
[3,3]
6
output:

input:
[1,2,3,4]
7
output:
```

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

If you want to add local testing support for a new language, please refer to [#112](https://github.com/j178/leetgo/issues/112).

Before submitting a PR, please run `golangci-lint run --fix` to fix lint errors.

## Credits

Here are some awesome projects that inspired me to create this project:

- https://github.com/EndlessCheng/codeforces-go
- https://github.com/clearloop/leetcode-cli
- https://github.com/budougumi0617/leetgode
- https://github.com/skygragon/leetcode-cli

Also thanks to [JetBrains](https://www.jetbrains.com/) for providing free licenses to support this project.

[![JetBrains Logo](https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.svg)](https://jb.gg/OpenSourceSupport)

[^1]: For Safari on macOS, you may need to grant `Full Disk Access` privilege to your terminal app which you would like to run `leetgo`.
