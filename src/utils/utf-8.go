package utils

import (
	"unicode"
	"unicode/utf8"
)

func IsValidUTF8(s string) (res bool) {
	defer func() {
		if recover() != nil {
			res = false
		}
	}()

	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError {
			return false
		}
		i += size
	}
	return true
}

func ChangeDisplaySafeUTF8(s string) (res string, safe bool) {
	defer func() {
		if recover() != nil {
			res = ""
			safe = false
		}
	}()

	if !IsValidUTF8(s) {
		return "", false
	}

	safe = true
	str := []rune(s)
	resRune := make([]rune, 0, len(str))

	for _, r := range str {
		if unicode.IsMark(r) || unicode.IsPunct(r) || unicode.IsSpace(r) || unicode.IsSymbol(r) {
			resRune = append(resRune, r)
		} else if unicode.IsControl(r) {
			safe = false
		} else if unicode.IsPrint(r) {
			resRune = append(resRune, r)
		} else {
			safe = false
		}
	}
	return string(resRune), true
}
