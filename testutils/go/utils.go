package goutils

import "strings"

func MustRead(s string, err error) string {
	if err != nil {
		panic(err)
	}
	return s
}

func JoinArray(s []string) string {
	return "[" + strings.Join(s, ",") + "]"
}
