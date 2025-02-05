package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/database"
	"github.com/SongZihuan/anonymous-message/src/email"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"github.com/SongZihuan/anonymous-message/src/iprate"
	"github.com/SongZihuan/anonymous-message/src/utils"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	"time"
)

const exp = 5 * time.Minute
const maxcount = 3

const msgtypetext = "text"
const atall = "@all"

type GetData struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Refer   string `json:"refer"`
}

type ReturnData struct {
	Code       int    `json:"code"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ErrMessage string `json:"error,omitempty"`
}

type WebhookText struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list"`
	MentionedMobileList []string `json:"mentioned_mobile_list"`
}

type ReqWebhookMsg struct {
	MsgType string      `json:"msgtype"`
	Text    WebhookText `json:"text"`
}

type RespWebhookMsg struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type SendError struct {
	Code    int
	Err     error
	Message string
}

func (s *SendError) Error() string {
	if s.Err == nil {
		return s.Message
	}

	return fmt.Sprintf("%s: %s", s.Message, s.Err.Error())
}

func HandlerMessage(c *gin.Context) {
	origin, ok := handlerOptions(c)
	if origin == "" || !ok {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	var JSON = func(code int, obj *ReturnData) {
		if !flagparser.Debug {
			obj.ErrMessage = ""
		}

		c.JSON(code, obj)
	}

	clientIP := c.ClientIP()
	if iprate.CheckIP(clientIP, exp) > maxcount {
		JSON(200, &ReturnData{
			Code:       -1,
			Success:    false,
			Message:    "留言太频繁，请稍后再留言。",
			ErrMessage: "IP限制",
		})
		return
	}

	var data GetData
	err := c.ShouldBindBodyWithJSON(&data)
	if err != nil {
		JSON(200, &ReturnData{
			Code:       -2,
			Success:    false,
			Message:    "留言信息错误，请通过邮件 songzihuan@song-zh.com 留言。",
			ErrMessage: err.Error(),
		})
		return
	}

	data.Name = strings.ReplaceAll(data.Name, "\r\n", "\n")
	data.Name = strings.TrimLeft(data.Name, "\n")
	data.Name = strings.TrimRight(data.Name, "\n")
	data.Name = strings.TrimSpace(data.Name)

	if len(data.Name) > 30 {
		JSON(200, &ReturnData{
			Code:       -3,
			Success:    false,
			Message:    "名字太长啦，请控制在25个字符以内。",
			ErrMessage: "名字超过30个字符",
		})
		return
	} else if data.Name == "" {
		data.Name = "匿名（Anonymous User）"
	}

	safeName, isSafeName := utils.ChangeDisplaySafeUTF8(data.Name)
	if safeName == "" {
		JSON(200, &ReturnData{
			Code:       -4,
			Success:    false,
			Message:    "留言存在编码（例如非UTF-8编码或包含控制符合）或不安全问题，留言失败。",
			ErrMessage: "UTF-8检查不通过",
		})
		return
	}

	data.Message = strings.ReplaceAll(data.Message, "\r\n", "\n")
	data.Message = strings.TrimLeft(data.Message, "\n")
	data.Message = strings.TrimRight(data.Message, "\n")
	data.Message = strings.TrimSpace(data.Message)

	if data.Message == "" {
		JSON(200, &ReturnData{
			Code:       -5,
			Success:    false,
			Message:    "留言消息不能为空哦。",
			ErrMessage: "消息为空",
		})
		return
	} else if len(data.Message) > 200 {
		JSON(200, &ReturnData{
			Code:       -6,
			Success:    false,
			Message:    "留言信息太长了，服务器只能接接受150个字符呢！",
			ErrMessage: "消息超过200个字符",
		})
		return
	}

	safeMsg, isSafeMsg := utils.ChangeDisplaySafeUTF8(data.Message)
	if safeMsg == "" {
		JSON(200, &ReturnData{
			Code:       -7,
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
			Message:    "留言信息错误，请通过邮件 songzihuan@song-zh.com 留言。",
			ErrMessage: "Refer超过50个字符",
		})
		return
	}

	safeRefer, isSafeRefer := utils.ChangeDisplaySafeUTF8(data.Refer)
	if safeRefer == "" || !isSafeRefer {
		JSON(200, &ReturnData{
			Code:       -10,
			Success:    false,
			Message:    "留言信息错误，请通过邮件 songzihuan@song-zh.com 留言。",
			ErrMessage: "Refer不安全",
		})
		return
	}

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		JSON(200, &ReturnData{
			Code:       -11,
			Success:    false,
			Message:    "留言信息错误，请通过邮件 songzihuan@song-zh.com 留言。",
			ErrMessage: "获取地区失败",
		})
		return
	}

	now := time.Now().In(loc)
	mailID := utils.GetMailID(safeName, safeMsg, safeRefer, origin, c.Request.Host, now)

	var msgBuilder strings.Builder
	msgBuilder.WriteString(fmt.Sprintf("邮件ID: %s\n", mailID))
	msgBuilder.WriteString(fmt.Sprintf("接收时间: %s %s\n", now.Format("2006-01-02 15:04:05"), now.Location().String()))

	msgBuilder.WriteString(fmt.Sprintf("站点：%s\n", safeRefer))

	msgBuilder.WriteString(fmt.Sprintf("Origin: %s\n", origin))
	msgBuilder.WriteString(fmt.Sprintf("Host: %s\n", c.Request.Host))

	msgBuilder.WriteString(fmt.Sprintf("名字：%s\n", safeName))
	if !isSafeName {
		msgBuilder.WriteString(fmt.Sprintf("注意：原名字可能包含不安全内容，已被删除（原名字长度：%d）\n", len(data.Name)))
	}

	msgBuilder.WriteString(fmt.Sprintf("IP地址：%s\n", clientIP))

	if !isSafeMsg {
		msgBuilder.WriteString(fmt.Sprintf("注意：消息可能包含不安全内容，已被删除（消息原长度：%d）\n", len(data.Message)))
	}

	msgBuilder.WriteString(fmt.Sprintf("消息长度：%d\n", len(safeMsg)))
	msgBuilder.WriteString(fmt.Sprintf("---消息开始---\n%s\n---消息结束---", safeMsg))

	msg := msgBuilder.String()

	vxchan := make(chan bool, 2)
	emailchan := make(chan bool, 2)

	go func(mailID string, name string, content string, refer string, origin string, host string, clientIP string, t time.Time, message string, vxchan chan bool, emailchan chan bool) {
		defer func() {
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

		err := DataBase(mailID, name, content, refer, origin, host, clientIP, t, message)
		if err != nil {
			fmt.Printf("数据库提交消息出现错误: %s", err.Error())
			vxchan <- false
			emailchan <- false
			return
		}

		vxchan <- true
		emailchan <- true
	}(mailID, safeName, safeMsg, safeRefer, origin, c.Request.Host, clientIP, now, msg, vxchan, emailchan)

	go func(msg string, vxchan chan bool) {
		err := func() (err error) {
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

			err = WechatRobot(msg)
			if err != nil {
				fmt.Printf("企业微信发送消息出现错误: %s\n", err.Error())
				return err
			}

			return nil
		}()

		func(vxErr error) {
			defer close(vxchan)
			if <-vxchan {
				_ = database.UpdateWxRobotSendMsg(mailID, vxErr)
			}
		}(err)
	}(msg, vxchan)

	go func(msg string, origin string, refer string, now time.Time, emailchan chan bool) {
		err := func() (err error) {
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

			err = Email(msg, origin, refer, now)
			if err != nil {
				fmt.Printf("邮件发送消息出现错误: %s\n", err.Error())
				return err
			}

			return nil
		}()

		func(emailErr error) {
			defer close(emailchan)
			if <-emailchan {
				_ = database.UpdateEmailSendMsg(mailID, emailErr)
			}
		}(err)
	}(msg, origin, safeRefer, now, emailchan)

	if isSafeMsg {
		JSON(200, &ReturnData{
			Code:       0,
			Success:    true,
			Message:    "留言成功！",
			ErrMessage: "",
		})
	} else {
		JSON(200, &ReturnData{
			Code:       1,
			Success:    true,
			Message:    "留言存在编码（例如非UTF-8编码或包含控制符合）或不安全问题，留言信息已被处理，留言成功！",
			ErrMessage: "",
		})
	}
}

func DataBase(mailID string, name string, content string, refer string, origin string, host string, clientIP string, t time.Time, message string) error {
	err := database.SaveMail(mailID, name, content, refer, origin, host, clientIP, t, message)
	if err != nil {
		return err
	}

	return nil
}

func Email(msg string, origin string, refer string, t time.Time) (err error) {
	if flagparser.SMTPAddress == "" || flagparser.SMTPUser == "" || flagparser.SMTPRecipient == "" {
		return nil
	}

	if refer != "" && origin != "" && refer != origin {
		err = email.Send(fmt.Sprintf("站点: %s（Origin: %s）", refer, origin), msg, t)
	} else if refer != "" {
		err = email.Send(fmt.Sprintf("站点: %s", refer), msg, t)
	} else if origin != "" {
		err = email.Send(fmt.Sprintf("站点 Origin: %s", origin), msg, t)
	} else {
		return &SendError{
			Code:    -1,
			Message: "Refer和Origin为空",
			Err:     nil,
		}
	}

	if err != nil {
		return &SendError{
			Code:    -1,
			Message: "邮件发送异常",
			Err:     err,
		}
	}

	return nil
}

func WechatRobot(msg string) error {
	if flagparser.Webhook == "" {
		return nil
	}

	if len([]byte(msg)) >= 2048 {
		return &SendError{
			Code:    -1,
			Message: "消息太长，超过企业微信限制",
			Err:     nil,
		}
	}

	webhookData, err := json.Marshal(ReqWebhookMsg{
		MsgType: msgtypetext,
		Text: WebhookText{
			Content:             msg,
			MentionedMobileList: []string{atall},
		},
	})
	if err != nil {
		return &SendError{
			Code:    -1,
			Message: "编码请求结构体为json错误",
			Err:     err,
		}
	}

	resp, err := http.Post(flagparser.Webhook, "application/json", bytes.NewBuffer(webhookData))
	if err != nil {
		return &SendError{
			Code:    -2,
			Message: "提交POST请求错误",
			Err:     err,
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return &SendError{
			Code:    -3,
			Message: "读取Body错误",
			Err:     err,
		}
	}

	var respWebhook RespWebhookMsg
	err = json.Unmarshal(respData, &respWebhook)
	if err != nil {
		return &SendError{
			Code:    -4,
			Message: "将body解析成json错误",
			Err:     err,
		}
	}

	if respWebhook.ErrCode != 0 {
		return &SendError{
			Code:    -5,
			Message: fmt.Sprintf("企业微信报告错误 (错误码: %d): %s", respWebhook.ErrCode, respWebhook.ErrMsg),
			Err:     nil,
		}
	}

	return nil
}
