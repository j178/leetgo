package goutils

func MustRead(s string, err error) string {
	if err != nil {
		panic(err)
	}
	return s
}
