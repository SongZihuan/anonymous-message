package messageutils

import (
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/database"
	"strings"
	"time"
)

// WriteMessageStdHeader formats and appends a standard header to the given message builder.
// It includes message type, mail ID, and reception time.
// Parameters:
//
//	msgBuilder: *strings.Builder - The builder to which the header will be appended.
//	msgType: string - The type of the message, e.g., database.MsgTypeWebsite.
//	mailID: string - The unique identifier for the mail.
//	t: time.Time - The time of receiving the message.
func WriteMessageStdHeader(msgBuilder *strings.Builder, msgType database.MsgType, mailID string, t time.Time) {
	msgBuilder.WriteString(fmt.Sprintf("【电子信箱】"))
	msgBuilder.WriteString(fmt.Sprintf("类型：%s\n", msgType))
	msgBuilder.WriteString(fmt.Sprintf("信件ID: %s\n", mailID))
	msgBuilder.WriteString(fmt.Sprintf("接收时间: %s %s\n", t.Format("2006-01-02 15:04:05"), t.Location().String()))
}

func WriteSNMessageStdHeader(msgBuilder *strings.Builder, msgType database.MsgType, mailID string, t time.Time) {
	if msgType != database.MsgTypeSystem {
		panic("not a system message")
	}

	msgBuilder.WriteString(fmt.Sprintf("【电子信箱-系统通知】"))
	msgBuilder.WriteString(fmt.Sprintf("信件ID: %s\n", mailID))
	msgBuilder.WriteString(fmt.Sprintf("接收时间: %s %s\n", t.Format("2006-01-02 15:04:05"), t.Location().String()))
}
