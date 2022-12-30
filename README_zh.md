# Leetgo

Best LeetCode friend for geek.

---

[![CI](https://github.com/j178/leetgo/actions/workflows/ci.yaml/badge.svg)](https://github.com/j178/leetgo/actions/workflows/ci.yaml)

中文 | [English](./README.md)

Leetgo is a command line tool that generates skeleton code for LeetCode questions. You can run and debug test cases locally with your favorite IDE.
Then you can submit your code to LeetCode directly.

**This project is in its early development stage, and anything is likely to change.**

## Main features

- Search for and view a question by its ID or slug.
- Generate skeleton code and testing code for a question.
- Run test cases on your local machine.
- Generate contest questions just in time.

## Install `leetgo`

You can download the latest binary from the [release page](https://github.com/j178/leetgo/releases).

### go install

```shell
go install github.com/j178/leetgo@latest
```

### brew install

```shell
brew install j178/tap/leetgo
```

## Supported languages

- Golang
- Python
- Rust

### Planning

- Java
- C++

## Usage
<!-- BEGIN USAGE -->
```
Usage:
  leetgo [command]

Available Commands:
  init                    Init a leetcode workspace
  new                     Generate a new question
  today                   Generate the question of today
  info                    Show question info
  test                    Run question test cases
  submit                  Submit solution
  contest                 Generate contest questions
  update                  Update local questions DB

Flags:
  -v, --version   version for leetgo

Use "leetgo [command] --help" for more information about a command.
```
<!-- END USAGE -->

## Config file
<!-- BEGIN CONFIG -->
```yaml
# Use Chinese language
cn: true
# LeetCode configuration
leetcode:
  # LeetCode site
  site: https://leetcode.cn
go:
  # Enable Go generator
  enable: false
  # Output directory for Go files
  out_dir: go
  # Generate separate package for each question
  separate_package: true
  # Filename template for Go files
  filename_template: ""
python:
  # Enable Python generator
  enable: false
  # Output directory for Python files
  out_dir: python
```
<!-- END CONFIG -->

## Credits

- https://github.com/EndlessCheng/codeforces-go
- https://github.com/clearloop/leetcode-cli
- https://github.com/budougumi0617/leetgode
