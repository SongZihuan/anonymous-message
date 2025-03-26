package sender

import (
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/database"
	"github.com/SongZihuan/anonymous-message/src/sender/internal"
	"time"
)

func SNDataBase(mailID string, subject string, content string, t time.Time) error {
	err := database.SaveSNMail(mailID, subject, content, t)
	if err != nil {
		return &internal.SendError{
			Code:    -1,
			Message: "数据库错误",
			Err:     err,
		}
	}

	return nil
}

func SNWechatRobot(msg string) (string, error) {
	internal.WeChatRobotLock.Lock()
	defer internal.WeChatRobotLock.Unlock()

	return internal.WechatRobotSystemNotifyToSelf(msg)
}

func SNEmail(subject string, msg string, t time.Time) (smtpID string, err error) {
	if subject != "" {
		smtpID, err = internal.EmailSendToSelf(fmt.Sprintf("系统通知：%s", subject), msg, t)
	} else {
		return smtpID, &internal.SendError{
			Code:    -1,
			Message: "subject 为空",
			Err:     nil,
		}
	}

	if err != nil {
		return smtpID, err
	}

	return smtpID, nil
}
