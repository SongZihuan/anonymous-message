package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func GetAMMailID(name string, email string, msg string, refer string, origin string, host string, t time.Time) string {
	text := fmt.Sprintf("AM-%s\n%s\n%s\n%s\n%s\n%s\n%d", name, email, msg, refer, origin, host, t.Unix())
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GetIMAPMailID(messageID string, sender string, from string, to string, replyTo string, subject string, content string, date time.Time, t time.Time) string {
	text := fmt.Sprintf("IMAP-%s\n%s\n%s\n%s\n%s\n%s\n%s\n%d\n%d", messageID, sender, from, to, replyTo, subject, content, date.Unix(), t.Unix())
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GetSNMailID(subject string, content string, t time.Time) string {
	text := fmt.Sprintf("SYSTEM-NOTIFY-%s\n%s\n%d", subject, content, t.Unix())
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
