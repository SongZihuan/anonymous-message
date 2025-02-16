package systemnotify

import (
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/database"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"github.com/SongZihuan/anonymous-message/src/messageutils"
	"github.com/SongZihuan/anonymous-message/src/sender"
	"github.com/SongZihuan/anonymous-message/src/utils"
	"strings"
	"time"
)

func SendNotify(notifySubject string, notifyContent string) {
	vxchan := make(chan bool, 2)
	emailchan := make(chan bool, 2)

	notifyTime := time.Now().In(flagparser.TimeZone())
	notifyMailID := utils.GetSNMailID(notifySubject, notifyContent, notifyTime)

	go func(mailID string, subject string, content string, t time.Time, vxchan chan bool, emailchan chan bool) {
		defer func() {
			defer func() {
				_ = recover() // 防止chan关闭
			}()

			// 兜底
			vxchan <- false
			emailchan <- false
		}()

		defer func() {
			if r := recover(); r != nil {
				if _err, ok := r.(error); ok {
					fmt.Printf("数据库提交消息出现致命错误: %s\n", _err.Error())
				} else {
					fmt.Printf("数据库提交消息出现致命错误（非error）: %v\n", r)
				}
			}
		}()

		err := sender.SNDataBase(mailID, subject, content, t)
		if err != nil {
			fmt.Printf("数据库提交消息出现错误: %s", err.Error())
			vxchan <- false
			emailchan <- false
			return
		}

		vxchan <- true
		emailchan <- true
	}(notifyMailID, notifySubject, notifyContent, notifyTime, vxchan, emailchan)

	go func(mailID string, subject string, content string, now time.Time, vxchan chan bool) {
		defer close(vxchan)

		messageStr, err := func() (messageStr string, err error) {
			defer func() {
				if r := recover(); r != nil {
					if _err, ok := r.(error); ok {
						fmt.Printf("企业微信发送消息出现致命错误: %s\n", _err.Error())
						if err != nil {
							err = _err
							return
						}
					} else {
						fmt.Printf("企业微信发送消息出现致命错误（非error）: %v\n", r)
						if err != nil {
							err = fmt.Errorf("%v", r)
							return
						}
					}
				}
			}()

			var msgBuilder strings.Builder
			// 标准头部
			messageutils.WriteSNMessageStdHeader(&msgBuilder, "System-Notify - 系统提示", mailID, now)

			msgBuilder.WriteString(fmt.Sprintf("主题: %s\n", subject))
			msgBuilder.WriteString(fmt.Sprintf("日期: %s %s\n", now.Format("2006-01-02 15:04:05"), now.Location().String()))
			msgBuilder.WriteString(fmt.Sprintf("消息长度：%d\n", len(content)))
			msgBuilder.WriteString(fmt.Sprintf("---消息开始---\n%s\n---消息结束---", content))
			messageStr = msgBuilder.String()

			err = sender.SNWechatRobot(messageStr)
			if err != nil {
				fmt.Printf("企业微信发送消息出现错误: %s\n", err.Error())
				return "", err
			}

			return messageStr, nil
		}()

		func(vxErr error) {
			if <-vxchan {
				_ = database.UpdateSNWxRobotSendMsg(mailID, messageStr, vxErr)
			}
		}(err)
	}(notifyMailID, notifySubject, notifyContent, notifyTime, vxchan)

	go func(mailID string, subject string, content string, now time.Time, emailchan chan bool) {
		defer close(emailchan)

		messageStr, smtpID, err := func() (messageStr string, smtpID string, err error) {
			defer func() {
				if r := recover(); r != nil {
					if _err, ok := r.(error); ok {
						fmt.Printf("邮件发送消息出现致命错误: %s\n", _err.Error())
						if err != nil {
							err = _err
							return
						}
					} else {
						fmt.Printf("邮件发送消息出现致命错误（非error）: %v\n", r)
						if err != nil {
							err = fmt.Errorf("%v", r)
							return
						}
					}
				}
			}()

			var msgBuilder strings.Builder
			// 标准头部
			messageutils.WriteSNMessageStdHeader(&msgBuilder, "System-Notify - 系统提示", mailID, now)

			msgBuilder.WriteString(fmt.Sprintf("主题: %s\n", subject))
			msgBuilder.WriteString(fmt.Sprintf("日期: %s %s\n", now.Format("2006-01-02 15:04:05"), now.Location().String()))
			msgBuilder.WriteString(fmt.Sprintf("消息长度：%d\n", len(content)))
			msgBuilder.WriteString(fmt.Sprintf("---消息开始---\n%s\n---消息结束---", content))
			messageStr = msgBuilder.String()

			smtpID, err = sender.SNEmail(subject, messageStr, now)
			if err != nil {
				fmt.Printf("邮件发送消息出现错误: %s\n", err.Error())
				return "", smtpID, err
			}

			return messageStr, smtpID, nil
		}()

		func() {
			if <-emailchan {
				_ = database.UpdateSNEmailSendMsg(mailID, messageStr, smtpID, err)
			}
		}()
	}(notifyMailID, notifySubject, notifyContent, notifyTime, emailchan)
}
