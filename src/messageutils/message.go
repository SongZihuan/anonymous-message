package messageutils

import (
	"fmt"
	"strings"
	"time"
)

// WriteMessageStdHeader formats and appends a standard header to the given message builder.
// It includes message type, mail ID, and reception time.
// Parameters:
//
//	msgBuilder: *strings.Builder - The builder to which the header will be appended.
//	msgType: string - The type of the message, e.g., "AM - 匿名留言".
//	mailID: string - The unique identifier for the mail.
//	t: time.Time - The time of receiving the message.
func WriteMessageStdHeader(msgBuilder *strings.Builder, msgType string, mailID string, t time.Time) {
	msgBuilder.WriteString(fmt.Sprintf("【电子信箱】"))
	msgBuilder.WriteString(fmt.Sprintf("类型：%s\n", msgType))
	msgBuilder.WriteString(fmt.Sprintf("邮件ID（MailID）: %s\n", mailID))
	msgBuilder.WriteString(fmt.Sprintf("接收时间: %s %s\n", t.Format("2006-01-02 15:04:05"), t.Location().String()))
}

func WriteSNMessageStdHeader(msgBuilder *strings.Builder, msgType string, mailID string, t time.Time) {
	msgBuilder.WriteString(fmt.Sprintf("【电子信箱-通知系统】"))
	//msgBuilder.WriteString(fmt.Sprintf("类型：%s\n", msgType))
	msgBuilder.WriteString(fmt.Sprintf("邮件ID（MailID）: %s\n", mailID))
	msgBuilder.WriteString(fmt.Sprintf("接收时间: %s %s\n", t.Format("2006-01-02 15:04:05"), t.Location().String()))

	_ = msgType // msgType 暂时不使用
}
