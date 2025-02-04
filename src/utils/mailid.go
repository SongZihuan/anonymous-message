package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func GetMailID(name string, msg string, refer string, origin string, host string, t time.Time) string {
	text := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%d", name, msg, refer, origin, host, t.Unix())
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
