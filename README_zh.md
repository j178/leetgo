**中文 | [English](./README.md)**

# Leetgo

[![Go Report Card](https://goreportcard.com/badge/github.com/j178/leetgo)](https://goreportcard.com/report/github.com/j178/leetgo)
[![CI](https://github.com/j178/leetgo/actions/workflows/ci.yaml/badge.svg)](https://github.com/j178/leetgo/actions/workflows/ci.yaml)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://makeapullrequest.com)
[![Discord](https://img.shields.io/discord/1069106479744962582?label=discord&logo=discord)](https://discord.gg/bHsEwQQj9m)
[![Twitter Follow](https://img.shields.io/twitter/follow/niceoe)](https://twitter.com/niceoe)


`leetgo` 是一个命令行工具，帮助你管理 LeetCode 代码、简化常用的 LeetCode 操作，让解题更轻松~

`leetgo` 可以自动生成题目描述、样例代码，最特别的是还可以生成测试代码，让你的代码可以在本地运行、测试，你可以使用喜欢的 Debugger 来调试代码中的问题。

`leetgo` 还支持竞赛模式，自动等待比赛的开始时间，第一时间为你生成所有比赛题目，并且可以一键提交所有题目，让你的排名更进一步。

[![asciicast](https://asciinema.org/a/7R2lnZj7T0hEuJ49SH2lZ04NG.svg)](https://asciinema.org/a/7R2lnZj7T0hEuJ49SH2lZ04NG)

## 主要特性

- 自动为题目生成描述、样例代码、测试代码
- 通过模板引擎自定义配置生成的代码文件，支持对代码做预处理
- 支持本地测试，可以使用 Debugger 调试代码
- 自动等待并及时生成竞赛题目，一键提交所有题目
- 同时支持 leetcode.com (美国站) 和 leetcode.cn (中国站)
- 自动从浏览器中读取 LeetCode 的 Cookie，无需手动提供
- 自动在你喜欢的编辑器中打开生成的代码文件

## 编程语言支持

`leetgo` 可以为大多数语言生成样例代码，以及为部分语言生成本地测试代码。

本地测试意味着你可以在你的机器上运行你的代码，输入测试样例比对结果，你可以使用 Debugger 来单步调试你的代码，更容易的找出代码中的问题。

本地测试需要为每一种语言做单独的适配，所以目前仅支持部分语言(其实只支持 Go)，下表是目前的支持情况：

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
其他热门语言的支持都在计划中，如果你有兴趣的话，欢迎加入我们👏🏻

## 安装

你可以直接从 [release 页面](https://github.com/j178/leetgo/releases) 下载最新的可执行程序，添加可执行权限、加入 `PATH` 后使用。

### 使用 `go install`
 
```shell
git clone git@github.com:j178/leetgo.git
cd leetgo && go install
```

### macOS/Linux 使用 [HomeBrew](https://brew.sh/)

```shell
brew install j178/tap/leetgo
```

### Windows 使用 [Scoop](https://scoop.sh/)

```shell
scoop bucket add j178 https://github.com/j178/scoop-bucket.git
scoop install j178/leetgo
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

### 题目标志符 `qid`

许多 `leetgo` 命令都依赖 `qid` 来定位 LeetCode 题目。`qid` 是 `leetgo` 定义的一种简化的题目标志符，目的是让指定一个题目更简单，支持多种形式：

```shell
leetgo pick two-sum          # two-sum 是题目的 slug，是最准确的 qid
leetgo pick 1                # 1 是题目的 ID
leetgo pick today            # today 表示今天的每日一题
leetgo contest weekly100     # weekly100 表示第100场周赛
leetgo test last             # last 表示最近一个生成的题目
leetgo test weekly100/1      # weekly100/1 表示第100场周赛的第一个题目
leetgo submit b100/2         # b100/2 表示第100场双周赛的第二个题目
leetgo submit w99/           # w99 表示第99场周赛的所有题目 (必须要保留末尾的斜杠，否则不会识别为周赛题目)
leetgo test last/1           # last/1 表示最近生成的比赛的第一个题目
leetgo test last/            # last/ 表示最近生成的比赛的所有题目 (必须要保留末尾的斜杠)
```

## 配置说明

`leetgo` 使用两级配置结构：全局配置和项目配置。

全局配置位于 `~/.config/leetgo/config.yaml`，项目配置是项目根目录中的 `leetgo.yaml` 文件。 他们都是在 `leetgo init` 初始化过程中生成的。

项目配置会覆盖全局配置中的相同配置。通常使用全局配置作为默认的配置，然后在各个项目中调整 `leetgo.yaml` 来自定义项目中的配置。

下面是一个完整配置的展示：

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

## LeetCode 支持情况

`leetgo` 使用 LeetCode 的 GraphQL API 来获取题目和提交代码，`leetgo` 需要 LeetCode 的 Cookie 来代替你做这些事情。

有三种方式为 `leetgo` 提供认证:

- 从浏览器中直接读取。
  
  这是最方便的方法，也是默认的行为。目前支持 Chrome，FireFox，Safari[^1]，Edge。

  如果你需要其他浏览器的支持，请告诉我们~

  ```yaml
  leetcode:
    credentials:
      from: browser
  ```

- 在配置文件中提供 Cookie
  
  你需要打开 LeetCode 页面，从浏览器的 DevTools 中获取 `LEETCODE_SESSION` 和 `csrftoken` 这两个 Cookie 的值。

  ```yaml
  leetcode:
    credentials:
      from: cookies
      session: xxx
      csrftoken: xx
  ```

- 在配置文件中提供 用户名和密码

  在配置密码前，你需要使用 `leetgo config encrypt` 来加密你的密码，`leetgo` **禁止**在配置文件中使用明文密码。

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

**注意**: 不推荐使用用户名密码的认证方式, 而且 `leetcode.com` (美国站) 也不支持用户名密码登录.

## 进阶用法

1. template 相关
  
    `leetgo` 的配置中有许多支持 Go template，如果你熟悉 Go template 语法的话，可以配置出更加个性化的文件名和代码模板。

2. Blocks

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

3. Script

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

提交前请使用 `golangci-lint run --fix` 来修复代码格式问题。

## 致谢

在 `leetgo` 的开发过程中，下面这些项目为我提供了许多灵感和参考，感谢他们 :heart:

- https://github.com/EndlessCheng/codeforces-go
- https://github.com/clearloop/leetcode-cli
- https://github.com/budougumi0617/leetgode
- https://github.com/skygragon/leetcode-cli

[^1]: 使用 Safari 的用户注意，你需要赋予使用 `leetgo` 的终端 App `全盘访问`的权限。
