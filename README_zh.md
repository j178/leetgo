**简体中文 | [English](./README.md)**

# Leetgo

[![Go Report Card](https://goreportcard.com/badge/github.com/j178/leetgo)](https://goreportcard.com/report/github.com/j178/leetgo)
[![CI](https://github.com/j178/leetgo/actions/workflows/ci.yaml/badge.svg)](https://github.com/j178/leetgo/actions/workflows/ci.yaml)
[![GitHub downloads](https://img.shields.io/github/downloads/j178/leetgo/total)](https://github.com/j178/leetgo/releases)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://makeapullrequest.com)
[![Twitter Follow](https://img.shields.io/twitter/follow/niceoe)](https://twitter.com/niceoe)

`leetgo` 是一个命令行工具，帮助你管理 LeetCode 代码、简化常用的 LeetCode 操作，让解题更轻松~

`leetgo` 可以自动生成题目描述、样例代码，最特别的是还可以生成测试代码，让你的代码可以在本地运行、测试，你可以使用喜欢的 Debugger 来调试代码中的问题。

`leetgo` 还支持竞赛模式，自动等待比赛的开始时间，第一时间为你生成所有比赛题目，并且可以一键提交所有题目，让你的排名更进一步。

## 快速开始

1. [安装 leetgo](#安装)
2. 创建一个项目: `leetgo init -t <us or cn> -l <lang>`
3. 编辑配置文件: `leetgo.yaml`
4. 选择一个问题: `leetgo pick <id or name or today>`
5. 测试你的代码: `leetgo test last -L`
6. 提交你的代码: `leetgo submit last`

你可以用一行命令实现测试并提交: `leetgo test last -L -s`

你可以在你最喜欢的编辑器中修改代码: `leetgo edit last`

## Demo

![demo](https://github.com/j178/leetgo/assets/10510431/8eaee981-a1f7-4b40-b9df-5af3c72daf26)

## 主要特性

- 自动为题目生成描述、样例代码、测试代码
- 通过模板引擎自定义配置生成的代码文件，支持对代码做预处理
- 支持本地测试，可以使用 Debugger 调试代码
- 自动等待并及时生成竞赛题目，一键提交所有题目
- 同时支持 leetcode.com (美国站) 和 leetcode.cn (中国站)
- 自动从浏览器中读取 LeetCode 的 Cookie，无需手动提供
- 自动在你喜欢的编辑器中打开生成的代码文件
- 使用 OpenAI 发现并自动修复你代码中问题 (Experimental)

## 编程语言支持

`leetgo` 可以为大多数语言生成样例代码，以及为部分语言生成本地测试代码。

以 Go 语言为例，`leetgo pick 257` 会生成如下代码：

```go
// 省略一些代码...
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

这是一个完整的可运行的程序，你可以直接运行它，输入测试样例，比对结果。`leetgo test -L` 会自动按照 `testcases.txt` 中的 case 运行这个程序，并且比对结果。

本地测试意味着你可以在你的机器上运行你的代码，输入测试样例比对结果，你可以使用 Debugger 来单步调试你的代码，更容易的找出代码中的问题。

本地测试需要为每一种语言做单独的适配，所以目前仅支持部分语言，下表是目前的支持情况：

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

如果你有兴趣，欢迎加入我们支持更多语言👏🏻

## 安装

你可以直接从 [release 页面](https://github.com/j178/leetgo/releases) 下载最新的可执行程序，添加可执行权限、加入 `PATH` 后使用。

### macOS/Linux 使用 [HomeBrew](https://brew.sh/)

```shell
brew install j178/tap/leetgo
```

### Windows 使用 [Scoop](https://scoop.sh/)

```shell
scoop bucket add j178 https://github.com/j178/scoop-bucket.git
scoop install j178/leetgo
```

### ArchLinux 使用 [AUR](https://aur.archlinux.org/packages/leetgo-bin/)

```shell
yay -S leetgo-bin
```

### macOS/Linux 使用脚本安装

```shell
curl -fsSL https://raw.githubusercontent.com/j178/leetgo/master/scripts/install.sh | bash
```

### 使用 `go install` 从源码安装

```shell
go install github.com/j178/leetgo@latest
```

## 使用
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

### 题目标志符 `qid`

许多 `leetgo` 命令都依赖 `qid` 来定位 LeetCode 题目。`qid` 是 `leetgo` 定义的一种简化的题目标志符，目的是让指定一个题目更简单，支持多种形式：

```shell
leetgo pick two-sum          # two-sum 是题目的 slug，是最准确的 qid
leetgo pick 1                # 1 是题目的 ID
leetgo pick today            # today 表示今天的每日一题
leetgo pick yesterday        # `yesterday` 表示昨天的每日一题
leetgo pick today-1          # `today-1` 表示昨天的每日一题，与 `yesterday` 一样. `today-2`, `today-3` 等同理。
leetgo contest weekly100     # weekly100 表示第100场周赛
leetgo test last             # last 表示最近一个生成的题目
leetgo test weekly100/1      # weekly100/1 表示第100场周赛的第一个题目
leetgo submit b100/2         # b100/2 表示第100场双周赛的第二个题目
leetgo submit w99/           # w99 表示第99场周赛的所有题目 (必须要保留末尾的斜杠，否则不会识别为周赛题目)
leetgo test last/1           # last/1 表示最近生成的比赛的第一个题目
leetgo test last/            # last/ 表示最近生成的比赛的所有题目 (必须要保留末尾的斜杠)
```

## 配置说明

> [!WARNING]
> 从 `v1.4` 开始，`leetgo` 不再读取全局的 `~/.config/leetgo/config.yaml` 文件，请将所有配置都放到项目的 `leetgo.yaml` 文件中。

`leetgo init` 会在当前目录生成一个 `leetgo.yaml` 文件，这个文件包含了 `leetgo` 的所有配置，你可以根据自己的需要修改这个文件。

`leetgo.yaml` 所在的目录会被认为是一个 `leetgo` 项目的根目录，`leetgo` 会在这个目录下生成所有的代码文件。`leetgo` 会在当前目录中查找 `leetgo.yaml` 文件，如果没有找到，会向上递归查找，直到找到一个 `leetgo.yaml` 文件或者到达文件系统的根目录。

下面是一个完整配置的展示：

<details>
<summary>Click to expand</summary>

<!-- BEGIN CONFIG -->
```yaml
# Your name
author: Bob
# Language of the question description: 'zh' (Simplified Chinese) or 'en' (English).
language: zh
code:
  # Language of code generated for questions: go, cpp, python, java... 
  # (will be overridden by command line flag -l/--lang).
  lang: go
  # The default template to generate filename (without extension), e.g. {{.Id}}.{{.Slug}}
  # Available attributes: Id, Slug, Title, Difficulty, Lang, SlugIsMeaningful
  # (Most questions have descriptive slugs, but some consist of random characters. The SlugIsMeaningful boolean indicates whether a slug is meaningful.)
  # Available functions: lower, upper, trim, padWithZero, toUnderscore, group.
  filename_template: '{{ .Id | padWithZero 4 }}{{ if .SlugIsMeaningful }}.{{ .Slug }}{{ end }}'
  # Generate question description into a separate question.md file, otherwise it will be embed in the code file.
  separate_description_file: true
  # Default modifiers for all languages.
  modifiers:
    - name: removeUselessComments
  go:
    # Base directory to put generated questions, defaults to the language slug, e.g. go, python, cpp.
    out_dir: go
    # Functions that modify the generated code.
    modifiers:
      - name: removeUselessComments
      - name: changeReceiverName
      - name: addNamedReturn
      - name: addMod
  python3:
    # Base directory to put generated questions, defaults to the language slug, e.g. go, python, cpp.
    out_dir: python
    # Path to the python executable that creates the venv.
    executable: python3
  cpp:
    # Base directory to put generated questions, defaults to the language slug, e.g. go, python, cpp.
    out_dir: cpp
    # C++ compiler
    cxx: g++
    # C++ compiler flags (our Leetcode I/O library implementation requires C++17).
    cxxflags: -O2 -std=c++17
  rust:
    # Base directory to put generated questions, defaults to the language slug, e.g. go, python, cpp.
    out_dir: rust
  java:
    # Base directory to put generated questions, defaults to the language slug, e.g. go, python, cpp.
    out_dir: java
leetcode:
  # LeetCode site, https://leetcode.com or https://leetcode.cn
  site: https://leetcode.cn
  # Credentials to access LeetCode.
  credentials:
    # How to provide credentials: browser, cookies, password or none.
    from: browser
    # Browsers to get cookies from: chrome, safari, edge or firefox. If empty, all browsers will be tried. Only used when 'from' is 'browser'.
    browsers: []
contest:
  # Base directory to put generated contest questions.
  out_dir: contest
  # Template to generate filename of the question.
  filename_template: '{{ .ContestShortSlug }}/{{ .Id }}{{ if .SlugIsMeaningful }}.{{ .Slug }}{{ end }}'
  # Open the contest page in browser after generating.
  open_in_browser: true
# Editor settings to open generated files.
editor:
  # Use a predefined editor: vim, vscode, goland
  # Set to 'none' to disable, set to 'custom' to provide your own command and args.
  use: none
  # Custom command to open files.
  command: ""
  # Arguments to your custom command.
  # String contains {{.CodeFile}}, {{.TestFile}}, {{.DescriptionFile}}, {{.TestCasesFile}} will be replaced with corresponding file path.
  # {{.Folder}} will be substituted with the output directory.
  # {{.Files}} will be substituted with the list of all file paths.
  args: ""
```
<!-- END CONFIG -->
</details>

## LeetCode 认证

`leetgo` 使用 LeetCode 的 GraphQL API 来获取题目和提交代码，`leetgo` 需要 LeetCode 的 Cookie 来代替你做这些事情。

有三种方式为 `leetgo` 提供认证:

- 从浏览器中直接读取。

  这是最方便的方法，也是默认的行为。目前支持 Chrome，FireFox，Safari[^1]，Edge。

  ```yaml
  leetcode:
    credentials:
      from: browser
  ```

- 手动提供 Cookie
  
  你需要打开 LeetCode 页面，从浏览器的 DevTools 中获取 `LEETCODE_SESSION` 和 `csrftoken` 这两个 Cookie 的值，设置为 `LEETCODE_SESSION` 和 `LEETCODE_CSRFTOKEN` 环境变量。如果你在使用 `leetcode.com`, 你还需要设置 `LEETCODE_CFCLEARANCE` 为 `cf_clearance` cookie 的值。

  ```yaml
  leetcode:
    credentials:
      from: cookies
  ```

- 提供 LeetCode CN 的用户名和密码，设置 `LEETCODE_USERNAME` 和 `LEETCODE_PASSWORD` 环境变量。

  ```yaml
  leetcode:
    credentials:
      from: password
  ```

> [!NOTE]
> 不推荐使用用户名密码的认证方式, 而且 `leetcode.com` (美国站) 也不支持用户名密码登录.

## 进阶用法

### `testcases.txt` 相关

`leetgo` 会自动为你生成 `testcases.txt` 文件，这个文件包含了所有测试用例，你可以在这个文件中添加自己的测试用例，`leetgo test` 会自动读取这个文件中的测试用例来作为程序的输入。

当你尚不清楚用例的输出时，你可以将 `output:` 部分留空。当执行 `leetgo test` 时，`leetgo` 自动将远程输出的正确结果填充到 `output:` 部分。示例：

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

### template 相关

`leetgo` 的配置中有许多支持 Go template，如果你熟悉 Go template 语法的话，可以配置出更加个性化的文件名和代码模板。

### Blocks

可以用 blocks 来自定义代码中的一些部分，目前支持的 block 有：
- header
- description
- title
- beforeMarker
- beforeCode
- code
- afterCode
- afterMarker

示例：
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

### Script

`leetgo` 支持自定义一个 JavaScript 脚本来处理函数代码，示例：
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

如果你在使用中遇到了问题，可以设置环境变量 `DEBUG=1` 来启动 Debug 模式，然后再运行 `leetgo`，比如 `DEBUG=1 leetgo test last`。

Debug 模式下 `leetgo` 会输出详细的日志，请复制这些日志，并且附带 `leetgo config` 的输出，向我们提交一个 issue，这对于我们定位问题至关重要。

一些常见问题请参考 [Q&A](https://github.com/j178/leetgo/discussions/categories/q-a)。

## 欢迎贡献代码

欢迎大家参与这个项目的开发，如果你不知道如何开始，这些 [Good first issues](https://github.com/j178/leetgo/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) 是很好的起步点，
你也可以看看这些 [help wanted](https://github.com/j178/leetgo/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22) issues。

如果你想为一个新的语言添加本地测试的支持，请参考 [#112](https://github.com/j178/leetgo/issues/112)。

提交前请使用 `golangci-lint run --fix` 来修复代码格式问题。

## 致谢

在 `leetgo` 的开发过程中，下面这些项目为我提供了许多灵感和参考，感谢他们 :heart:

- https://github.com/EndlessCheng/codeforces-go
- https://github.com/clearloop/leetcode-cli
- https://github.com/budougumi0617/leetgode
- https://github.com/skygragon/leetcode-cli

也感谢 [JetBrains](https://www.jetbrains.com/) 为本项目提供的免费开源许可证。

[![JetBrains Logo](https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.svg)](https://jb.gg/OpenSourceSupport)

[^1]: 使用 Safari 的用户注意，你需要赋予使用 `leetgo` 的终端 App `全盘访问`的权限。
