package utils

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime/quotedprintable"
	"net/mail"
	"strings"
)

var NotUTF8Encoding = fmt.Errorf("not utf-8 encoding")
var NotStdEncoding = fmt.Errorf("not B or Q encoding")
var BadEncodingNotSafe = fmt.Errorf("bad encoding not safe")

func FormatEmailAddressToHumanStringJustNameMustSafe(ad *mail.Address) string {
	res, err := FormatEmailAddressToHumanStringJustNameSafe(ad)
	if err != nil || res == "" {
		if strings.Contains(strings.ToLower(ad.Name), "unsafe1") {
			return "unsafe2"
		}
		return "unsafe1"
	}

	return res
}

func FormatEmailAddressToHumanStringJustNameSafe(ad *mail.Address) (string, error) {
	res := FormatEmailAddressToHumanStringJustName(ad)
	res, _ = ChangeDisplaySafeUTF8(res)
	if res == "" {
		localPart, _, err := SplitEmailAddress(ad.Address)
		if err == nil {
			res, _ = ChangeDisplaySafeUTF8(localPart)
			if res == "" {
				return "", fmt.Errorf("unsafe")
			}

			return res, nil
		}

		res, _ = ChangeDisplaySafeUTF8(ad.Address)
		if res == "" {
			return "", fmt.Errorf("unsafe")
		}

		return res, nil
	}

	return res, nil
}

func FormatEmailAddressToHumanStringJustName(ad *mail.Address) string {
	name := ad.Name
	address := ad.Address

	if name == "" {
		localPart, _, err := SplitEmailAddress(address)
		if err == nil {
			return localPart
		}
		return address
	}

	_name, err := DecodeEmailEncodings(name)
	if err == nil {
		return _name
	}

	return name
}

func FormatEmailAddressToHumanStringMustSafe(ad *mail.Address) string {
	res, err := FormatEmailAddressToHumanStringSafe(ad)
	if err != nil || res == "" {
		if strings.Contains(strings.ToLower(ad.Address), "unsafe1@example.com") {
			return "unsafe2@example.com"
		}
		return "unsafe1@example.com"
	}

	return res
}

func FormatEmailAddressToHumanStringSafe(ad *mail.Address) (string, error) {
	res := FormatEmailAddressToHumanString(ad)
	res, _ = ChangeDisplaySafeUTF8(res)
	if res == "" {
		res, _ = ChangeDisplaySafeUTF8(ad.Address)
		if res == "" {
			return "", fmt.Errorf("unsafe")
		}

		return res, nil
	}

	return res, nil
}

func FormatEmailAddressToHumanString(ad *mail.Address) string {
	name := ad.Name
	address := ad.Address

	if name == "" {
		return address
	}

	_name, err := DecodeEmailEncodings(name)
	if err == nil {
		name = _name
	}

	res := fmt.Sprintf("%s <%s>", name, address)
	return res
}

func DecodeEmailEncodingsSafe(encodedStr string) (res string, isSafe bool, err error) {
	if encodedStr == "" {
		return encodedStr, true, nil
	}

	res, err = DecodeEmailEncodings(encodedStr)
	if err != nil {
		return "", false, err
	}

	res, isSafe = ChangeDisplaySafeUTF8(res)
	if res == "" {
		return "", false, BadEncodingNotSafe
	}

	return res, isSafe, nil
}

func DecodeEmailEncodings(encodedStr string) (decodedStr string, err error) {
	dest := strings.TrimSuffix(encodedStr, `"`)
	dest = strings.TrimPrefix(encodedStr, `"`)

	const prefix = `=?`
	const suffix = `?=`

	if !strings.HasPrefix(dest, prefix) && !strings.HasSuffix(dest, suffix) {
		return encodedStr, nil
	}

	const utf8 = prefix + `UTF-8?`

	if !strings.HasPrefix(strings.ToUpper(dest), `=?UTF-8?`) {
		return "", NotUTF8Encoding
	}

	const b = utf8 + `B?`
	const q = utf8 + `Q?`

	if strings.HasPrefix(strings.ToUpper(dest), b) {
		target := dest[len(b) : len(dest)-len(suffix)]

		res, err := base64.StdEncoding.DecodeString(target)
		if err != nil {
			return "", err
		}

		return string(res), nil
	} else if strings.HasPrefix(strings.ToUpper(dest), q) {
		target := dest[len(q) : len(dest)-len(suffix)]

		reader := quotedprintable.NewReader(strings.NewReader(target))
		res, err := io.ReadAll(reader)
		if err != nil {
			return "", err
		}

		return string(res), nil
	} else {
		return "", NotStdEncoding
	}
}

func SplitEmailAddress(email string) (string, string, error) {
	if !IsValidEmail(email) {
		return "", "", fmt.Errorf("not a valid email")
	}

	atIndex := strings.Index(email, "@")
	if atIndex == -1 {
		return "", "", fmt.Errorf("invalid email address: missing '@'")
	}

	localPart := email[:atIndex]
	domainPart := email[atIndex+1:]
	return localPart, domainPart, nil
}
