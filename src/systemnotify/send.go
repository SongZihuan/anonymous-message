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
	now := time.Now().In(flagparser.TimeZone())
	notifyMailID := utils.GetSNMailID(notifySubject, notifyContent, now)

	initchan := make(chan bool)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if _err, ok := r.(error); ok {
					fmt.Printf("数据库提交消息出现致命错误: %s\n", _err.Error())
				} else {
					fmt.Printf("数据库提交消息出现致命错误（非error）: %v\n", r)
				}
			}
		}()

		defer close(initchan)

		err := sender.SNDataBase(notifyMailID, notifySubject, notifyContent, now)
		if err != nil {
			fmt.Printf("数据库提交消息出现错误: %s\n", err.Error())
		}
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if _err, ok := r.(error); ok {
					fmt.Printf("企业微信发送消息出现致命错误: %s\n", _err.Error())
				} else {
					fmt.Printf("企业微信发送消息出现致命错误（非error）: %v\n", r)
				}
			}
		}()

		var msgBuilder strings.Builder
		// 标准头部
		messageutils.WriteSNMessageStdHeader(&msgBuilder, database.MsgTypeSystem, notifyMailID, now)

		msgBuilder.WriteString(fmt.Sprintf("主题: %s\n", notifySubject))
		msgBuilder.WriteString(fmt.Sprintf("日期: %s %s\n", now.Format("2006-01-02 15:04:05"), now.Location().String()))
		msgBuilder.WriteString(fmt.Sprintf("消息长度：%d\n", len(notifyContent)))
		msgBuilder.WriteString(fmt.Sprintf("---消息开始---\n%s\n---消息结束---", notifyContent))
		messageStr := msgBuilder.String()

		wxrobotID, err := sender.SNWechatRobot(messageStr)
		if err != nil {
			fmt.Printf("企业微信发送消息出现错误: %s\n", err.Error())
			return
		}

		_ = database.UpdateSNWxRobotSendMsg(notifyMailID, wxrobotID)
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if _err, ok := r.(error); ok {
					fmt.Printf("邮件发送消息出现致命错误: %s\n", _err.Error())
				} else {
					fmt.Printf("邮件发送消息出现致命错误（非error）: %v\n", r)
				}
			}
		}()

		var msgBuilder strings.Builder
		// 标准头部
		messageutils.WriteSNMessageStdHeader(&msgBuilder, database.MsgTypeSystem, notifyMailID, now)

		msgBuilder.WriteString(fmt.Sprintf("主题: %s\n", notifySubject))
		msgBuilder.WriteString(fmt.Sprintf("日期: %s %s\n", now.Format("2006-01-02 15:04:05"), now.Location().String()))
		msgBuilder.WriteString(fmt.Sprintf("消息长度：%d\n", len(notifyContent)))
		msgBuilder.WriteString(fmt.Sprintf("---消息开始---\n%s\n---消息结束---", notifyContent))
		messageStr := msgBuilder.String()

		smtpID, err := sender.SNEmail(notifySubject, messageStr, now)
		if err != nil {
			fmt.Printf("邮件发送消息出现错误: %s\n", err.Error())
		}

		_ = database.UpdateSNEmailSendMsg(notifyMailID, smtpID)
	}()
}
