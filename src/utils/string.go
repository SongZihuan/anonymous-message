package utils

import (
	"regexp"
	"strings"
	"unicode"
)

const BASE_CHAR = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStr(length int) string {
	bytes := []byte(BASE_CHAR)

	var result []byte
	for i := 0; i < length; i++ {
		result = append(result, bytes[Rand().Intn(len(bytes))])
	}

	return string(result)
}

func InvalidPhone(phone string) bool {
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

func IsValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

const NormalConsoleWidth = 80

func FormatTextToWidth(text string, width int) string {
	return FormatTextToWidthAndPrefix(text, 0, width)
}

func FormatTextToWidthAndPrefix(text string, prefixWidth int, overallWidth int) string {
	var result strings.Builder

	width := overallWidth - prefixWidth
	if width <= 0 {
		panic("bad width")
	}

	text = strings.ReplaceAll(text, "\r\n", "\n")

	for _, line := range strings.Split(text, "\n") {
		result.WriteString(strings.Repeat(" ", prefixWidth))

		if line == "" {
			result.WriteString("\n")
			continue
		}

		spaceCount := CountSpaceInStringPrefix(line) % width
		newLineLength := 0
		if spaceCount < 80 {
			result.WriteString(strings.Repeat(" ", spaceCount))
			newLineLength = spaceCount
		}

		for _, word := range strings.Fields(line) {
			if newLineLength+len(word) >= width {
				result.WriteString("\n")
				result.WriteString(strings.Repeat(" ", prefixWidth))
				newLineLength = 0
			}

			// 不是第一个词时，添加空格
			if newLineLength != 0 {
				result.WriteString(" ")
				newLineLength += 1
			}

			result.WriteString(word)
			newLineLength += len(word)
		}

		if newLineLength != 0 {
			result.WriteString("\n")
			newLineLength = 0
		}
	}

	return strings.TrimRight(result.String(), "\n")
}

func CountSpaceInStringPrefix(str string) int {
	var res int
	for _, r := range str {
		if r == ' ' {
			res += 1
		} else {
			break
		}
	}

	return res
}

func IsValidURLPath(path string) bool {
	if path == "" {
		return true
	} else if path == "/" {
		return false
	}

	pattern := `^\/[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;%=]+$`
	matched, _ := regexp.MatchString(pattern, path)
	return matched
}

func IsValidDomain(domain string) bool {
	pattern := `^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`
	matched, _ := regexp.MatchString(pattern, domain)
	return matched
}

func StringToOnlyPrint(str string) string {
	runeLst := []rune(str)
	res := make([]rune, 0, len(runeLst))

	for _, r := range runeLst {
		if unicode.IsPrint(r) {
			res = append(res, r)
		}
	}

	return string(res)
}

func IsGoodQueryKey(key string) bool {
	pattern := `^[a-zA-Z0-9\-._~]+$`
	matched, _ := regexp.MatchString(pattern, key)
	return matched
}

func IsValidHTTPHeaderKey(key string) bool {
	pattern := `^[a-zA-Z0-9!#$%&'*+.^_` + "`" + `|~-]+$`
	matched, _ := regexp.MatchString(pattern, key)
	return matched
}

func CompressSpaces(input string) string {
	var res strings.Builder

	lastIsSpace := false
	for _, r := range input {
		if r == ' ' {
			if lastIsSpace {
				continue
			} else {
				lastIsSpace = true
				res.WriteRune(r)
			}
		} else {
			lastIsSpace = false
			res.WriteRune(r)
		}
	}

	return res.String()
}

func CompressTTab(input string) string {
	var res strings.Builder

	lastIsTTab := false
	for _, r := range input {
		if r == '\t' {
			if lastIsTTab {
				continue
			} else {
				lastIsTTab = true
				res.WriteRune(r)
			}
		} else {
			lastIsTTab = false
			res.WriteRune(r)
		}
	}

	return res.String()
}

func Compress0xA0(input string) string {
	var res strings.Builder

	lastIs0xA0 := false
	for _, r := range input {
		if r == 0xA0 {
			if lastIs0xA0 {
				continue
			} else {
				lastIs0xA0 = true
				res.WriteRune(r)
			}
		} else {
			lastIs0xA0 = false
			res.WriteRune(r)
		}
	}

	return res.String()
}

func CompressSpacesGroup(input string) string {
	var res strings.Builder

	lastIsSpace := false
	for _, r := range input {
		if r == ' ' || r == 0xA0 || r == '\t' {
			if lastIsSpace {
				continue
			} else {
				lastIsSpace = true
				res.WriteRune(' ')
			}
		} else {
			lastIsSpace = false
			res.WriteRune(r)
		}
	}

	return res.String()
}

func CompressFormFeed(input string) string {
	var res strings.Builder

	lastIsFormFeed := false
	for _, r := range input {
		if r == '\f' {
			if lastIsFormFeed {
				continue
			} else {
				lastIsFormFeed = true
				res.WriteRune(r)
			}
		} else {
			lastIsFormFeed = false
			res.WriteRune(r)
		}
	}

	return res.String()
}

func Compress0x85(input string) string {
	var res strings.Builder

	lastIs0x85 := false
	for _, r := range input {
		if r == 0x85 {
			if lastIs0x85 {
				continue
			} else {
				lastIs0x85 = true
				res.WriteRune(r)
			}
		} else {
			lastIs0x85 = false
			res.WriteRune(r)
		}
	}

	return res.String()
}

func CompressEnter(input string) string {
	var res strings.Builder

	lastIsEnter := false
	for _, r := range input {
		if r == '\n' || r == '\r' {
			if lastIsEnter {
				continue
			} else {
				lastIsEnter = true
				res.WriteRune('\n')
			}
		} else {
			lastIsEnter = false
			res.WriteRune(r)
		}
	}

	return res.String()
}

func CompressVTab(input string) string {
	var res strings.Builder

	lastIsVTab := false
	for _, r := range input {
		if r == '\v' {
			if lastIsVTab {
				continue
			} else {
				lastIsVTab = true
				res.WriteRune(r)
			}
		} else {
			lastIsVTab = false
			res.WriteRune(r)
		}
	}

	return res.String()
}

func CompressEnterGroup(input string) string {
	var res strings.Builder

	lastIsEnter := false
	for _, r := range input {
		if r == '\n' || r == '\r' || r == 0x85 || r == '\f' || r == '\v' {
			if lastIsEnter {
				continue
			} else {
				lastIsEnter = true
				res.WriteRune('\n')
			}
		} else {
			lastIsEnter = false
			res.WriteRune(r)
		}
	}

	return res.String()
}

func Compress(input string) string {
	var res strings.Builder

	lastIsSpace := false
	for _, r := range input {
		if unicode.IsSpace(r) {
			if lastIsSpace {
				continue
			} else {
				lastIsSpace = true
				res.WriteRune(' ')
			}
		} else {
			lastIsSpace = false
			res.WriteRune(r)
		}
	}

	return res.String()
}

func CompressAuto(input string, target int) (string, bool) {
	dest := input

	if len(dest) <= target {
		return dest, true
	}

	if dest = CompressSpaces(dest); len(dest) <= target {
		return dest, true
	}

	if dest = CompressEnter(dest); len(dest) <= target {
		return dest, true
	}

	if dest = CompressTTab(dest); len(dest) <= target {
		return dest, true
	}

	if dest = CompressFormFeed(dest); len(dest) <= target {
		return dest, true
	}

	if dest = CompressVTab(dest); len(dest) <= target {
		return dest, true
	}

	if dest = Compress0xA0(dest); len(dest) <= target {
		return dest, true
	}

	if dest = Compress0x85(dest); len(dest) <= target {
		return dest, true
	}

	if dest = CompressSpacesGroup(dest); len(dest) <= target {
		return dest, true
	}

	if dest = CompressEnterGroup(dest); len(dest) <= target {
		return dest, true
	}

	if dest = Compress(dest); len(dest) <= target {
		return dest, true
	}

	return dest, false
}

func IsEmptyLine(str string) bool {
	for _, r := range str {
		if !unicode.IsSpace(r) {
			return false
		}
	}

	return true
}
