package utils

import (
	"encoding/hex"
	"strings"
	"unicode/utf16"
	"unsafe"
)

// BytesToString converts byte slice to string.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToBytes converts string to byte slice.
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// CondenseEmptyLines condenses multiple consecutive empty lines in a string to a single empty line,
// while preserving non-empty lines.
func CondenseEmptyLines(s string) string {
	lines := strings.Split(s, "\n")
	var filtered []string
	for i := 0; i < len(lines); i++ {
		if i == 0 || lines[i] != "" || lines[i-1] != "" {
			filtered = append(filtered, lines[i])
		}
	}
	return strings.Join(filtered, "\n")
}

func EnsureTrailingNewline(s string) string {
	if s == "" || s[len(s)-1] != '\n' {
		return s + "\n"
	}
	return s
}

var (
	subscripts = map[string]string{
		"0": "\u2080",
		"1": "\u2081",
		"2": "\u2082",
		"3": "\u2083",
		"4": "\u2084",
		"5": "\u2085",
		"6": "\u2086",
		"7": "\u2087",
		"8": "\u2088",
		"9": "\u2089",
		"i": "\u1d62",
		"j": "\u2c7c",
		"a": "\u2090",
		"e": "\u2091",
		"o": "\u2092",
		"x": "\u2093",
		"y": "\u1d67",
		"h": "\u2095",
		"k": "\u2096",
		"l": "\u2097",
		"m": "\u2098",
		"n": "\u2099",
		"p": "\u209a",
		"s": "\u209b",
		"t": "\u209c",
	}
	superscripts = map[string]string{
		"0": "\u2070",
		"1": "\u00b9",
		"2": "\u00b2",
		"3": "\u00b3",
		"4": "\u2074",
		"5": "\u2075",
		"6": "\u2076",
		"7": "\u2077",
		"8": "\u2078",
		"9": "\u2079",
		"i": "\u2071",
		"n": "\u207f",
	}
	subReplace = func() *strings.Replacer {
		args := make([]string, 0, len(subscripts)*2)
		for k, v := range subscripts {
			args = append(args, k, v)
		}
		return strings.NewReplacer(args...)
	}()
	supReplace = func() *strings.Replacer {
		args := make([]string, 0, len(superscripts)*2)
		for k, v := range superscripts {
			args = append(args, k, v)
		}
		return strings.NewReplacer(args...)
	}()
)

func ReplaceSubscript(s string) string {
	return subReplace.Replace(s)
}

func ReplaceSuperscript(s string) string {
	return supReplace.Replace(s)
}

func DecodeRawUnicodeEscape(s string) string {
	var buf strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+5 < len(s) && s[i+1] == 'u' {
			b16, _ := hex.DecodeString(s[i+2 : i+6])
			value := uint16(b16[0])<<8 + uint16(b16[1])
			chr := utf16.Decode([]uint16{value})[0]
			buf.WriteRune(chr)
			i += 5
		} else {
			buf.WriteByte(s[i])
		}
	}
	return buf.String()
}
