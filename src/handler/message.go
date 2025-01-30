package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		JSON(200, &ReturnData{
			Code:       -8,
			Success:    false,
			Message:    "留言信息错误，请通过邮件 songzihuan@song-zh.com 留言。",
			ErrMessage: "缺少Refer",
		})
		return
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

	var msgBuilder strings.Builder
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
	if len([]byte(msg)) >= 2048 {
		JSON(200, &ReturnData{
			Code:       -11,
			Success:    false,
			Message:    "留言信息太长了，缩短一点吧！",
			ErrMessage: "消息超过微信限制长度",
		})
		return
	}

	webhookData, err := json.Marshal(ReqWebhookMsg{
		MsgType: msgtypetext,
		Text: WebhookText{
			Content:             msg,
			MentionedMobileList: []string{atall},
		},
	})
	if err != nil {
		JSON(200, &ReturnData{
			Code:       -12,
			Success:    false,
			Message:    "留言信息错误，请通过邮件 songzihuan@song-zh.com 留言。",
			ErrMessage: err.Error(),
		})
		return
	}

	resp, err := http.Post(flagparser.Webhook, "application/json", bytes.NewBuffer(webhookData))
	if err != nil {
		JSON(200, &ReturnData{
			Code:       -13,
			Success:    false,
			Message:    "留言信息错误，请通过邮件 songzihuan@song-zh.com 留言。",
			ErrMessage: err.Error(),
		})
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		JSON(200, &ReturnData{
			Code:       -14,
			Success:    false,
			Message:    "留言信息错误，请通过邮件 songzihuan@song-zh.com 留言。",
			ErrMessage: err.Error(),
		})
		return
	}

	var respWebhook RespWebhookMsg
	err = json.Unmarshal(respData, &respWebhook)
	if err != nil {
		JSON(200, &ReturnData{
			Code:       -14,
			Success:    false,
			Message:    "留言信息错误，请通过邮件 songzihuan@song-zh.com 留言。",
			ErrMessage: err.Error(),
		})
		return
	}

	if respWebhook.ErrCode != 0 {
		JSON(200, &ReturnData{
			Code:       -14,
			Success:    false,
			Message:    "留言信息错误，请通过邮件 songzihuan@song-zh.com 留言。",
			ErrMessage: fmt.Sprintf("Webhook error (code: %d): %s", respWebhook.ErrCode, respWebhook.ErrMsg),
		})
		return
	}

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
