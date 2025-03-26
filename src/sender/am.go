package sender

import (
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/database"
	"github.com/SongZihuan/anonymous-message/src/sender/internal"
	"time"
)

func AMDataBase(mailID string, name string, email string, content string, refer string, origin string, host string, clientIP string, t time.Time) error {
	err := database.SaveAMMail(mailID, name, email, content, refer, origin, host, clientIP, t)
	if err != nil {
		return &internal.SendError{
			Code:    -1,
			Message: "数据库错误",
			Err:     err,
		}
	}

	return nil
}

func AMWechatRobot(msg string) (string, error) {
	internal.WeChatRobotLock.Lock()
	defer internal.WeChatRobotLock.Unlock()

	return internal.WechatRobotToSelf(msg)
}

func AMWechatRobotFile(msg string, file string) (wxrobotID string, fileID string, err error) {
	internal.WeChatRobotLock.Lock()
	defer internal.WeChatRobotLock.Unlock()

	wxrobotID, err = internal.WechatRobotToSelf(msg)
	if err != nil {
		return wxrobotID, "", err
	}

	return internal.WechatRobotFileToSelf(file, wxrobotID)
}

func AMEmail(msg string, origin string, refer string, t time.Time) (smtpID string, err error) {
	if refer != "" && origin != "" && refer != origin {
		smtpID, err = internal.EmailSendToSelf(fmt.Sprintf("站点: %s（Origin: %s）", refer, origin), msg, t)
	} else if refer != "" {
		smtpID, err = internal.EmailSendToSelf(fmt.Sprintf("站点: %s", refer), msg, t)
	} else if origin != "" {
		smtpID, err = internal.EmailSendToSelf(fmt.Sprintf("站点 Origin: %s", origin), msg, t)
	} else {
		return smtpID, &internal.SendError{
			Code:    -1,
			Message: "Refer和Origin为空",
			Err:     nil,
		}
	}

	if err != nil {
		return smtpID, err
	}

	return smtpID, nil
}
