package handler

import (
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/database"
	"github.com/SongZihuan/anonymous-message/src/email/imapserver"
	"github.com/SongZihuan/anonymous-message/src/email/smtpserver"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"github.com/SongZihuan/anonymous-message/src/messageutils"
	"github.com/SongZihuan/anonymous-message/src/reqrate"
	"github.com/SongZihuan/anonymous-message/src/sender"
	"github.com/SongZihuan/anonymous-message/src/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/mail"
	"strings"
	"time"
)

// RateExp: 计算周期
// RateMaxCount: 周期内最大发件数量
const (
	RateExp      time.Duration = 5 * time.Minute
	RateMaxCount int64         = 3
)

const DefaultName = "匿名（Anonymous User）"

type GetData struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
	Refer   string `json:"refer"`
}

type ReturnData struct {
	Code       int    `json:"code"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ErrMessage string `json:"error,omitempty"`
}

func HandlerMessage(c *gin.Context) {
	origin, ok := handlerOptions(c)
	if origin == "" || !ok {
		c.AbortWithStatus(http.StatusForbidden)
		return
	} else if len(origin) > 50 {
		if flagparser.Debug {
			_, _ = c.Writer.WriteString(fmt.Sprintf("Origin 头太长（不能超过50字符，现在字符数: %d）: %s", len(origin), origin))
		}
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	host := c.Request.Host
	if len(host) > 50 {
		if flagparser.Debug {
			_, _ = c.Writer.WriteString(fmt.Sprintf("Host 头太长（不能超过50字符，现在字符数: %d）: %s", len(host), host))
		}
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	var JSON = func(code int, obj *ReturnData) {
		if !flagparser.Debug {
			obj.ErrMessage = ""
		}

		if !strings.HasSuffix(obj.Message, "。") && !strings.HasSuffix(obj.Message, "！") {
			obj.Message += "。"
		}

		c.JSON(code, obj)
	}

	var data GetData
	err := c.ShouldBindBodyWithJSON(&data)
	if err != nil {
		JSON(200, &ReturnData{
			Code:       -1,
			Success:    false,
			Message:    "留言信息错误，请通过电子邮件留言。",
			ErrMessage: err.Error(),
		})
		return
	}

	data.Email = strings.ReplaceAll(data.Email, "\r\n", "\n")
	data.Email = strings.TrimLeft(data.Email, "\n")
	data.Email = strings.TrimRight(data.Email, "\n")
	data.Email = strings.TrimSpace(data.Email)

	if data.Email != "" {
		if utils.IsValidEmail(data.Email) {
			senderAddress := imapserver.DefaultFromAddress
			address := &mail.Address{
				Name:    "",
				Address: data.Email,
			}

			if senderAddress != nil {
				JSON = func(code int, obj *ReturnData) {
					if !flagparser.Debug {
						obj.ErrMessage = ""
					}

					if !strings.HasSuffix(obj.Message, "。") && !strings.HasSuffix(obj.Message, "！") {
						obj.Message += "。"
					}

					c.JSON(code, obj)

					if obj.Code >= 0 {
						return
					}

					msg := strings.TrimRight(obj.Message, "。")
					msg = strings.TrimRight(msg, "！")

					_, _ = smtpserver.SendErrorMsg("信件拒收通知", "", address, senderAddress, address, msg)
				}
			}
		} else {
			JSON(200, &ReturnData{
				Code:       -2,
				Success:    false,
				Message:    "邮箱错误。",
				ErrMessage: "邮箱错误",
			})
			return
		}
	}

	var IPRateLimit = false
	var EmailRateLimit = false

	clientIP := c.ClientIP()
	if reqrate.CheckHttpReqIP(clientIP, RateExp) > RateMaxCount {
		IPRateLimit = true
	}

	if data.Email != "" && reqrate.CheckMailStringAddressRate(data.Email, RateExp) > RateMaxCount {
		EmailRateLimit = true
	}

	if IPRateLimit && EmailRateLimit {
		JSON(200, &ReturnData{
			Code:       -3,
			Success:    false,
			Message:    "留言太频繁，请稍后再留言。",
			ErrMessage: "IP和Email限制",
		})
		return
	} else if IPRateLimit {
		JSON(200, &ReturnData{
			Code:       -3,
			Success:    false,
			Message:    "留言太频繁，请稍后再留言。",
			ErrMessage: "IP限制",
		})
		return
	} else if EmailRateLimit {
		EmailRateLimit = true
		JSON(200, &ReturnData{
			Code:       -3,
			Success:    false,
			Message:    "留言太频繁，请稍后再留言。",
			ErrMessage: "邮箱限制",
		})
		return
	}

	data.Name = strings.ReplaceAll(data.Name, "\r\n", "\n")
	data.Name = strings.TrimLeft(data.Name, "\n")
	data.Name = strings.TrimRight(data.Name, "\n")
	data.Name = strings.TrimSpace(data.Name)

	isAnonymous := false
	safeName := DefaultName
	isSafeName := true

	if len(data.Name) > 30 {
		JSON(200, &ReturnData{
			Code:       -4,
			Success:    false,
			Message:    "名字太长啦，请控制在25个字符以内。",
			ErrMessage: "名字超过30个字符",
		})
		return
	} else if data.Name == "" {
		isAnonymous = true
		data.Name = DefaultName
		safeName = data.Name
		isSafeName = true
	} else {
		safeName, isSafeName = utils.ChangeDisplaySafeUTF8(data.Name)
		if safeName == "" {
			JSON(200, &ReturnData{
				Code:       -5,
				Success:    false,
				Message:    "留言存在编码（例如非UTF-8编码或包含控制符合）或不安全问题，留言失败。",
				ErrMessage: "UTF-8检查不通过",
			})
			return
		}

		if data.Email != "" && imapserver.DefaultFromAddress != nil {
			senderAddress := imapserver.DefaultFromAddress
			address := &mail.Address{
				Name:    safeName,
				Address: data.Email,
			}

			JSON = func(code int, obj *ReturnData) {
				if !flagparser.Debug {
					obj.ErrMessage = ""
				}

				if !strings.HasSuffix(obj.Message, "。") && !strings.HasSuffix(obj.Message, "！") {
					obj.Message += "。"
				}

				c.JSON(code, obj)

				if obj.Code >= 0 {
					return
				}

				msg := strings.TrimRight(obj.Message, "。")
				msg = strings.TrimRight(obj.Message, "！")

				_, _ = smtpserver.SendErrorMsg("信件拒收通知", "", address, senderAddress, address, msg)
			}
		}
	}

	data.Message = strings.ReplaceAll(data.Message, "\r\n", "\n")
	data.Message = strings.TrimLeft(data.Message, "\n")
	data.Message = strings.TrimRight(data.Message, "\n")
	data.Message = strings.TrimSpace(data.Message)

	if data.Message == "" {
		JSON(200, &ReturnData{
			Code:       -6,
			Success:    false,
			Message:    "留言消息不能为空哦。",
			ErrMessage: "消息为空",
		})
		return
	} else if len(data.Message) > 200 {
		_message, ok := utils.CompressAuto(data.Message, 200)
		if ok {
			data.Message = _message
		} else {
			JSON(200, &ReturnData{
				Code:       -7,
				Success:    false,
				Message:    "留言信息太长了，服务器只能接接受150个字符呢！",
				ErrMessage: "消息超过200个字符",
			})
			return
		}
	}

	safeMsg, isSafeMsg := utils.ChangeDisplaySafeUTF8(data.Message)
	if safeMsg == "" {
		JSON(200, &ReturnData{
			Code:       -8,
			Success:    false,
			Message:    "留言存在编码（例如非UTF-8编码或包含控制符合）或不安全问题，留言失败。",
			ErrMessage: "UTF-8检查不通过",
		})
		return
	}

	if data.Refer == "" {
		data.Refer = origin
	} else if len(data.Refer) >= 50 {
		JSON(200, &ReturnData{
			Code:       -9,
			Success:    false,
			Message:    "留言信息错误，请通过电子邮件留言。",
			ErrMessage: "Refer超过50个字符",
		})
		return
	}

	safeRefer, isSafeRefer := utils.ChangeDisplaySafeUTF8(data.Refer)
	if safeRefer == "" || !isSafeRefer {
		JSON(200, &ReturnData{
			Code:       -10,
			Success:    false,
			Message:    "留言信息错误，请通过电子邮件留言。",
			ErrMessage: "Refer不安全",
		})
		return
	}

	now := time.Now().In(flagparser.TimeZone())
	mailID := utils.GetAMMailID(safeName, data.Email, safeMsg, safeRefer, origin, c.Request.Host, now)

	vxchan := make(chan bool, 2)
	emailchan := make(chan bool, 2)
	thankchan := make(chan bool, 2)

	go func(mailID string, name string, email string, content string, refer string, origin string, host string, clientIP string, t time.Time, vxchan chan bool, emailchan chan bool, thankchan chan bool) {
		defer func() {
			defer func() {
				_ = recover() // 防止chan关闭
			}()

			// 兜底
			vxchan <- false
			emailchan <- false
			thankchan <- false
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

		err := sender.AMDataBase(mailID, name, email, content, refer, origin, host, clientIP, t)
		if err != nil {
			fmt.Printf("数据库提交消息出现错误: %s", err.Error())
			vxchan <- false
			emailchan <- false
			thankchan <- false
			return
		}

		vxchan <- true
		emailchan <- true
		thankchan <- true
	}(mailID, safeName, data.Email, safeMsg, safeRefer, origin, c.Request.Host, clientIP, now, vxchan, emailchan, thankchan)

	go func(vxchan chan bool) {
		defer close(vxchan)

		msg, err := func() (msg string, err error) {
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
			messageutils.WriteMessageStdHeader(&msgBuilder, "AM - 匿名留言", mailID, now)

			msgBuilder.WriteString(fmt.Sprintf("站点：%s\n", safeRefer))

			msgBuilder.WriteString(fmt.Sprintf("Origin: %s\n", origin))
			msgBuilder.WriteString(fmt.Sprintf("Host: %s\n", c.Request.Host))
			msgBuilder.WriteString(fmt.Sprintf("IP地址：%s\n", clientIP))

			msgBuilder.WriteString(fmt.Sprintf("名字：%s\n", safeName))
			if !isSafeName {
				msgBuilder.WriteString(fmt.Sprintf("注意：原名字可能包含不安全内容，已被删除（原名字长度：%d）\n", len(data.Name)))
			}

			if isAnonymous {
				msgBuilder.WriteString(fmt.Sprintf("是否匿名：是\n"))
			} else {
				msgBuilder.WriteString(fmt.Sprintf("是否匿名：否\n"))
			}

			if data.Email != "" {
				msgBuilder.WriteString(fmt.Sprintf("邮箱：%s\n", data.Email))
			} else {
				msgBuilder.WriteString(fmt.Sprintf("邮箱：未预留\n"))
			}

			if !isSafeMsg {
				msgBuilder.WriteString(fmt.Sprintf("注意：消息可能包含不安全内容，已被删除（消息原长度：%d）\n", len(data.Message)))
			}

			msgBuilder.WriteString(fmt.Sprintf("消息长度：%d\n", len(safeMsg)))

			const start = "---消息开始---\n"
			const stop = "\n---消息结束---"
			const splitTip = "以下消息的正文部分将开始分段发送\n"
			const splitAllTip = "以下新消息将开始分段发送\n"

			tmpMsg := msgBuilder.String()

			if len(tmpMsg)+len(start)+len(safeMsg)+len(stop) <= 2040 {
				msgBuilder.WriteString(fmt.Sprintf("%s%s%s", start, safeMsg, stop))

				msg = msgBuilder.String()

				err = sender.AMWechatRobot(msg)
				if err != nil {
					fmt.Printf("企业微信发送消息出现错误: %s\n", err.Error())
					return "", err
				}

				return msg, nil
			} else if len(tmpMsg)+len(splitTip) <= 2040 {
				msgBuilder.WriteString(splitTip)
				msg = msgBuilder.String()

				err = sender.AMWechatRobot(msg)
				if err != nil {
					fmt.Printf("企业微信发送消息出现错误: %s\n", err.Error())
					return "", err
				}

				sendMsg, err := sender.AMWechatRobotSplitMsg(safeMsg, mailID)
				if err != nil {
					fmt.Printf("企业微信发送消息出现错误: %s\n", err.Error())
					return "", err
				}

				msgBuilder.WriteString(sendMsg)
				msg = msgBuilder.String()

				return msg, nil
			} else {
				msgBuilder.WriteString(fmt.Sprintf("%s%s%s", start, safeMsg, stop))
				msg = splitAllTip + msgBuilder.String()

				sendMsg, err := sender.AMWechatRobotSplitMsg(msg, mailID)
				if err != nil {
					fmt.Printf("企业微信发送消息出现错误: %s\n", err.Error())
					return "", err
				}

				return sendMsg, nil
			}
		}()

		func(vxErr error) {
			if <-vxchan {
				_ = database.UpdateAMWxRobotSendMsg(mailID, msg, vxErr)
			}
		}(err)
	}(vxchan)

	go func(origin string, refer string, now time.Time, emailchan chan bool) {
		defer close(emailchan)

		msg, smtpID, err := func() (msg string, smtpID string, err error) {
			defer func() {
				if r := recover(); r != nil {
					if _err, ok := r.(error); ok {
						fmt.Printf("电子邮件发送消息出现致命错误: %s\n", _err.Error())
						if err != nil {
							err = _err
							return
						}
					} else {
						fmt.Printf("电子邮件发送消息出现致命错误（非error）: %v\n", r)
						if err != nil {
							err = fmt.Errorf("%v", r)
							return
						}
					}
				}
			}()

			var msgBuilder strings.Builder
			// 标准头部
			messageutils.WriteMessageStdHeader(&msgBuilder, "AM - 匿名留言", mailID, now)

			msgBuilder.WriteString(fmt.Sprintf("站点：%s\n", safeRefer))

			msgBuilder.WriteString(fmt.Sprintf("Origin: %s\n", origin))
			msgBuilder.WriteString(fmt.Sprintf("Host: %s\n", c.Request.Host))
			msgBuilder.WriteString(fmt.Sprintf("IP地址：%s\n", clientIP))

			msgBuilder.WriteString(fmt.Sprintf("名字：%s\n", safeName))
			if !isSafeName {
				msgBuilder.WriteString(fmt.Sprintf("注意：原名字可能包含不安全内容，已被删除（原名字长度：%d）\n", len(data.Name)))
			}

			if isAnonymous {
				msgBuilder.WriteString(fmt.Sprintf("是否匿名：是\n"))
			} else {
				msgBuilder.WriteString(fmt.Sprintf("是否匿名：否\n"))
			}

			if data.Email != "" {
				msgBuilder.WriteString(fmt.Sprintf("邮箱：%s\n", data.Email))
			} else {
				msgBuilder.WriteString(fmt.Sprintf("邮箱：未预留\n"))
			}

			if !isSafeMsg {
				msgBuilder.WriteString(fmt.Sprintf("注意：消息可能包含不安全内容，已被删除（消息原长度：%d）\n", len(data.Message)))
			}

			msgBuilder.WriteString(fmt.Sprintf("消息长度：%d\n", len(safeMsg)))
			msgBuilder.WriteString(fmt.Sprintf("---消息开始---\n%s\n---消息结束---", safeMsg))

			msg = msgBuilder.String()

			smtpID, err = sender.AMEmail(msg, origin, refer, now)
			if err != nil {
				fmt.Printf("电子邮件发送消息出现错误: %s\n", err.Error())
				return "", smtpID, err
			}

			return msg, smtpID, nil
		}()

		func() {
			if <-emailchan {
				_ = database.UpdateAMEmailSendMsg(mailID, msg, smtpID, err)
			}
		}()
	}(origin, safeRefer, now, emailchan)

	go func(name string, email string, isAnonymous bool, thankchan chan bool) {
		defer close(thankchan)

		var smtpID string
		var err error

		if imapserver.DefaultFromAddress != nil && email != "" {
			var address *mail.Address

			if isAnonymous {
				address = &mail.Address{
					Name:    "",
					Address: email,
				}
			} else {
				address = &mail.Address{
					Name:    name,
					Address: email,
				}
			}

			smtpID, err = smtpserver.SendThankMsg("我们已经收到你的信件啦！", "", address, imapserver.DefaultFromAddress, address)
		} else {
			smtpID = ""
			err = fmt.Errorf("未启用IMAP服务，或者对方未留下邮箱")
		}

		if <-thankchan {
			_ = database.UpdateAMThankEmailSendMsg(mailID, smtpID, err)
		}
	}(safeName, data.Email, isAnonymous, thankchan)

	if isSafeMsg {
		JSON(200, &ReturnData{
			Code:       1,
			Success:    true,
			Message:    "留言成功！",
			ErrMessage: "",
		})
	} else {
		JSON(200, &ReturnData{
			Code:       2,
			Success:    true,
			Message:    "留言存在编码（例如非UTF-8编码或包含控制符合）或不安全问题，留言信息已被处理，留言成功！",
			ErrMessage: "",
		})
	}
}
