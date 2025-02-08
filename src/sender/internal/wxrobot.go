package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"io"
	"net/http"
	"sync"
)

const (
	msgtypetext     = "text"
	msgtypemarkdown = "markdown"
)
const atall = "@all"

var WeChatRobotLock sync.Mutex

type WebhookText struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list"`
	MentionedMobileList []string `json:"mentioned_mobile_list"`
}

type WebhookMarkdown struct {
	Content string `json:"content"`
}

type ReqWebhookMsg struct {
	MsgType  string           `json:"msgtype"`
	Text     *WebhookText     `json:"text,omitempty"`
	Markdown *WebhookMarkdown `json:"markdown,omitempty"`
}

type RespWebhookMsg struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func WechatRobotToSelf(msg string) error {
	if flagparser.Webhook == "" || msg == "" {
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
		Text: &WebhookText{
			Content: msg,
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

func WechatRobotSystemNotifyToSelf(msg string) error {
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
		Text: &WebhookText{
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

func WechatRobotMarkdownToSelf(msg string) error {
	if flagparser.Webhook == "" {
		return nil
	}

	if len([]byte(msg)) >= 4096 {
		return &SendError{
			Code:    -1,
			Message: "消息太长，超过企业微信限制",
			Err:     nil,
		}
	}

	webhookData, err := json.Marshal(ReqWebhookMsg{
		MsgType: msgtypemarkdown,
		Markdown: &WebhookMarkdown{
			Content: msg,
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
