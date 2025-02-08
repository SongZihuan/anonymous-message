package sender

import (
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/database"
	"github.com/SongZihuan/anonymous-message/src/sender/internal"
	"github.com/SongZihuan/anonymous-message/src/utils"
	"strings"
	"time"
)

func AMDataBase(mailID string, name string, content string, refer string, origin string, host string, clientIP string, t time.Time) error {
	err := database.SaveAMMail(mailID, name, content, refer, origin, host, clientIP, t)
	if err != nil {
		return &internal.SendError{
			Code:    -1,
			Message: "数据库错误",
			Err:     err,
		}
	}

	return nil
}

func AMWechatRobot(msg string) error {
	internal.WeChatRobotLock.Lock()
	defer internal.WeChatRobotLock.Unlock()

	return internal.WechatRobotToSelf(msg)
}

func AMWechatRobotSplitMsg(msg string, mailID string) (string, error) {
	internal.WeChatRobotLock.Lock()
	defer internal.WeChatRobotLock.Unlock()

	if len(mailID) > 5 {
		mailID = mailID[:5]
	}

	prefixTag := fmt.Sprintf("【%s】", mailID)

	var msgBuilder strings.Builder
	var cup = prefixTag

	for _, line := range strings.Split(msg, "\n") {
		if utils.IsEmptyLine(line) {
			continue
		}

		line = line + "\n"

		// 原始条件：len(line) > 2040 || len(prefixTag) + len(line) > 2040
		//  1. line超过2040，他则无法加入cup中（因为会导致cup也超过2040），所以要先发送cup，然后分段发送line
		//  2. line没超过2040，但如果进入逻辑2，cup被发送，而prefixTag + line超过2040个字符，是不正确的。应该在第一个逻辑处
		//     发送cup，然后分段发送line。
		//  3. 至于逻辑3没什么好说的，line长度足够小，拼接进去即可
		// 合并逻辑：
		// 当 len(line) 大于 2040 时，len(prefixTag) + len(line) 必然也大于 2040，因此可以合并逻辑，变成如下所示：
		if len(prefixTag)+len(line) > 2040 {
			msgBuilder.WriteString(cup)
			err := internal.WechatRobotToSelf(cup)
			if err != nil {
				return "", err
			}

			for line != "" {
				var subLine = prefixTag

				if len(line) > 2040 {
					subLine = line[:2040-len(subLine)]
					line = line[2040-len(subLine):]
				} else {
					subLine = line
					line = ""
				}

				if utils.IsEmptyLine(subLine) {
					continue
				}

				msgBuilder.WriteString(subLine)
				err = internal.WechatRobotToSelf(subLine)
				if err != nil {
					return "", err
				}
			}

			cup = prefixTag
		} else if len(cup)+len(line) > 2040 {
			msgBuilder.WriteString(cup)
			err := internal.WechatRobotToSelf(cup)
			if err != nil {
				return "", err
			}

			cup = prefixTag + line
		} else {
			cup += line
		}
	}

	return msgBuilder.String(), nil
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
