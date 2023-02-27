package goutils

import (
	"bufio"
	"strings"
)

func ReadLine(r *bufio.Reader) string {
	if line, err := r.ReadString('\n'); err != nil {
		panic(err)
	} else {
		return line
	}
}

func JoinArray(s []string) string {
	return "[" + strings.Join(s, ",") + "]"
}
