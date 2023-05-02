package utils

import (
	"encoding/hex"
	"strings"
	"unicode"
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

func CamelToSnake(name string) string {
	var snakeStrBuilder strings.Builder

	for i, r := range name {
		if i > 0 && unicode.IsUpper(r) && !unicode.IsUpper([]rune(name)[i-1]) {
			snakeStrBuilder.WriteRune('_')
		}
		snakeStrBuilder.WriteRune(unicode.ToLower(r))
	}

	return snakeStrBuilder.String()
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
		"a": "\u2090",
		"e": "\u2091",
		"h": "\u2095",
		"i": "\u1d62",
		"j": "\u2c7c",
		"k": "\u2096",
		"l": "\u2097",
		"m": "\u2098",
		"n": "\u2099",
		"o": "\u2092",
		"p": "\u209a",
		"r": "\u1d63",
		"s": "\u209b",
		"t": "\u209c",
		"u": "\u1d64",
		"v": "\u1d65",
		"x": "\u2093",
		"y": "\u1d67",
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
		"a": "\u1D43",
		"b": "\u1D47",
		"c": "\u1D9C",
		"d": "\u1D48",
		"e": "\u1D49",
		"f": "\u1DA0",
		"g": "\u1D4D",
		"h": "\u02B0",
		"i": "\u2071",
		"j": "\u02B2",
		"k": "\u1D4F",
		"l": "\u02E1",
		"m": "\u1D50",
		"n": "\u207F",
		"o": "\u1D52",
		"p": "\u1D56",
		"q": "\u02A0",
		"r": "\u02B3",
		"s": "\u02E2",
		"t": "\u1D57",
		"u": "\u1D58",
		"v": "\u1D5B",
		"w": "\u02B7",
		"x": "\u02E3",
		"y": "\u02B8",
		"z": "\u1DBB",
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
