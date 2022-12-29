# Leetgo

Solve leetcode problems in your terminal

---

**This project is in its early development stage, and anything is likely to change.**

## Main features

- Search for and view a problem by its ID or slug.
- Generate template and testing code for a problem.
- Run test cases on your local machine.
- Generate contest questions just in time.

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
  init        Init a leetcode workspace
  new         Generate a new question
  today       Generate the question of today
  info        Show question info
  test        Run question test cases
  submit      Submit solution
  contest     Generate contest questions
  update      Update local questions DB

Flags:
  -v, --version   version for leetgo

Use "leetgo [command] --help" for more information about a command.
```
<!-- END USAGE -->

## Config file
<!-- BEGIN CONFIG -->
```yaml
cn: true
leetcode:
  site: https://leetcode.cn
go:
  enable: false
  out_dir: go
  separate_package: true
  filename_template: ""
python:
  enable: false
  out_dir: python
```
<!-- END CONFIG -->

## Credits

- https://github.com/EndlessCheng/codeforces-go
- https://github.com/clearloop/leetcode-cli
- https://github.com/budougumi0617/leetgode
