package internal

import (
	"github.com/SongZihuan/anonymous-message/src/emailserver/smtpserver"
	"time"
)

func EmailSendToSelf(subject string, msg string, t time.Time) (smtpID string, err error) {
	if subject != "" || msg == "" {
		smtpID, err = smtpserver.SendToSelf(subject, msg, t)
	} else {
		return smtpID, &SendError{
			Code:    -1,
			Message: "subject 或 from 为空",
			Err:     nil,
		}
	}

	if err != nil {
		return smtpID, &SendError{
			Code:    -1,
			Message: "邮件发送异常",
			Err:     err,
		}
	}

	return smtpID, nil
}
