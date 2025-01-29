package utils

import "strings"

func OriginClear(origin string) string {
	origin = strings.TrimSpace(origin)
	origin = strings.TrimRight(origin, "/")

	if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
		return ""
	}

	return origin
}
