package sender

import (
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/database"
	"github.com/SongZihuan/anonymous-message/src/sender/internal"
	"time"
)

func IMAPDataBase(mailID string, messageID string, sender string, from string, to string, replyTo string, subject string, content string, date time.Time, t time.Time) error {
	err := database.SaveIMAPMail(mailID, messageID, sender, from, to, replyTo, subject, content, date, t)
	if err != nil {
		return &internal.SendError{
			Code:    -1,
			Message: "数据库错误",
			Err:     err,
		}
	}

	return nil
}

func IMAPWechatRobot(msg string) (string, error) {
	internal.WeChatRobotLock.Lock()
	defer internal.WeChatRobotLock.Unlock()

	return internal.WechatRobotToSelf(msg)
}

func IMAPEmail(subject string, from string, msg string, t time.Time) (smtpID string, err error) {
	if subject != "" && from != "" {
		smtpID, err = internal.EmailSendToSelf(fmt.Sprintf("邮件：%s (%s)", subject, from), msg, t)
	} else if subject != "" {
		smtpID, err = internal.EmailSendToSelf(fmt.Sprintf("邮件：%s", subject), msg, t)
	} else if from != "" {
		smtpID, err = internal.EmailSendToSelf(fmt.Sprintf("来自 %s 的邮件", from), msg, t)
	} else {
		return smtpID, &internal.SendError{
			Code:    -1,
			Message: "subject 或 from 为空",
			Err:     nil,
		}
	}

	if err != nil {
		return smtpID, err
	}

	return smtpID, nil
}
