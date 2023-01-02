#!/bin/sh
set -e
rm -rf completions
mkdir completions
export CGO_ENABLED=0
for sh in bash zsh fish; do
	go run main.go completion "$sh" >"completions/leetgo.$sh"
done
