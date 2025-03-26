package internal

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/database"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	msgtypetext     = "text"
	msgtypemarkdown = "markdown"
	msgtypefile     = "file"
)

const atall = "@all"

const upload_media = "https://qyapi.weixin.qq.com/cgi-bin/webhook/upload_media"

const (
	upload_file  = "file"
	upload_voice = "voice"
)

var WeChatRobotLock sync.Mutex

var onceGetKey sync.Once
var key string

type WebhookText struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list"`
	MentionedMobileList []string `json:"mentioned_mobile_list"`
}

type WebhookMarkdown struct {
	Content string `json:"content"`
}

type WebhookFile struct {
	MediaID string `json:"media_id"`
}

type ReqWebhookMsg struct {
	MsgType  string           `json:"msgtype"`
	Text     *WebhookText     `json:"text,omitempty"`
	Markdown *WebhookMarkdown `json:"markdown,omitempty"`
	File     *WebhookFile     `json:"file,omitempty"`
}

type RespWebhookMsg struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type RespWebhookFile struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	Type      string `json:"type"`
	MediaID   string `json:"media_id"`
	CreatedAt string `json:"created_at"`
}

func WechatRobotToSelf(msg string) (wxrobotID string, err error) {
	if flagparser.Webhook == "" || msg == "" {
		return "", nil
	}

	now := time.Now()

	defer func() {
		if wxrobotID != "" {
			_ = database.UpdateWxRobotRecord(wxrobotID, err)
		}
	}()

	wxrobotID = getWxRobotID(flagparser.Webhook, msg, now)

	err = database.SaveWxRobotRecord(wxrobotID, flagparser.Webhook, msg, now)
	if err != nil {
		return "", err
	}

	if len([]byte(msg)) >= 2048 {
		return wxrobotID, &SendError{
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
		return wxrobotID, &SendError{
			Code:    -1,
			Message: "编码请求结构体为json错误",
			Err:     err,
		}
	}

	resp, err := http.Post(flagparser.Webhook, "application/json", bytes.NewBuffer(webhookData))
	if err != nil {
		return wxrobotID, &SendError{
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
		return wxrobotID, &SendError{
			Code:    -3,
			Message: "读取Body错误",
			Err:     err,
		}
	}

	var respWebhook RespWebhookMsg
	err = json.Unmarshal(respData, &respWebhook)
	if err != nil {
		return wxrobotID, &SendError{
			Code:    -4,
			Message: "将body解析成json错误",
			Err:     err,
		}
	}

	if respWebhook.ErrCode != 0 {
		return wxrobotID, &SendError{
			Code:    -5,
			Message: fmt.Sprintf("企业微信报告错误 (错误码: %d): %s", respWebhook.ErrCode, respWebhook.ErrMsg),
			Err:     nil,
		}
	}

	return wxrobotID, nil
}

func WechatRobotSystemNotifyToSelf(msg string) (wxrobotID string, err error) {
	if flagparser.Webhook == "" || msg == "" {
		return "", nil
	}

	now := time.Now()

	defer func() {
		if wxrobotID != "" {
			_ = database.UpdateWxRobotRecord(wxrobotID, err)
		}
	}()

	wxrobotID = getWxRobotID(flagparser.Webhook, msg, now)

	err = database.SaveWxRobotRecord(wxrobotID, flagparser.Webhook, msg, now)
	if err != nil {
		return wxrobotID, err
	}

	if len([]byte(msg)) >= 2048 {
		return wxrobotID, &SendError{
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
		return wxrobotID, &SendError{
			Code:    -1,
			Message: "编码请求结构体为json错误",
			Err:     err,
		}
	}

	resp, err := http.Post(flagparser.Webhook, "application/json", bytes.NewBuffer(webhookData))
	if err != nil {
		return wxrobotID, &SendError{
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
		return wxrobotID, &SendError{
			Code:    -3,
			Message: "读取Body错误",
			Err:     err,
		}
	}

	var respWebhook RespWebhookMsg
	err = json.Unmarshal(respData, &respWebhook)
	if err != nil {
		return wxrobotID, &SendError{
			Code:    -4,
			Message: "将body解析成json错误",
			Err:     err,
		}
	}

	if respWebhook.ErrCode != 0 {
		return wxrobotID, &SendError{
			Code:    -5,
			Message: fmt.Sprintf("企业微信报告错误 (错误码: %d): %s", respWebhook.ErrCode, respWebhook.ErrMsg),
			Err:     nil,
		}
	}

	return wxrobotID, nil
}

func WechatRobotMarkdownToSelf(msg string) (wxrobotID string, err error) {
	if flagparser.Webhook == "" || msg == "" {
		return "", nil
	}

	now := time.Now()

	defer func() {
		if wxrobotID != "" {
			_ = database.UpdateWxRobotRecord(wxrobotID, err)
		}
	}()

	wxrobotID = getWxRobotID(flagparser.Webhook, msg, now)

	err = database.SaveWxRobotRecord(wxrobotID, flagparser.Webhook, msg, now)
	if err != nil {
		return "", err
	}

	if len([]byte(msg)) >= 4096 {
		return wxrobotID, &SendError{
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
		return wxrobotID, &SendError{
			Code:    -1,
			Message: "编码请求结构体为json错误",
			Err:     err,
		}
	}

	resp, err := http.Post(flagparser.Webhook, "application/json", bytes.NewBuffer(webhookData))
	if err != nil {
		return wxrobotID, &SendError{
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
		return wxrobotID, &SendError{
			Code:    -3,
			Message: "读取Body错误",
			Err:     err,
		}
	}

	var respWebhook RespWebhookMsg
	err = json.Unmarshal(respData, &respWebhook)
	if err != nil {
		return wxrobotID, &SendError{
			Code:    -4,
			Message: "将body解析成json错误",
			Err:     err,
		}
	}

	if respWebhook.ErrCode != 0 {
		return wxrobotID, &SendError{
			Code:    -5,
			Message: fmt.Sprintf("企业微信报告错误 (错误码: %d): %s", respWebhook.ErrCode, respWebhook.ErrMsg),
			Err:     nil,
		}
	}

	return wxrobotID, nil
}

func WechatRobotFileToSelf(msg string, wxrobotID string) (_ string, fileID string, err error) {
	if flagparser.Webhook == "" || msg == "" {
		return "", "", nil
	}

	defer func() {
		if wxrobotID != "" {
			_ = database.UpdateWxRobotFileRecord(wxrobotID, fileID, err)
		}
	}()

	err = database.SaveWxRobotFileRecord(wxrobotID, msg)
	if err != nil {
		return "", "", err
	}

	fileID, err = uploadMsgFile(msg)
	if err != nil {
		return wxrobotID, fileID, &SendError{
			Code:    -1,
			Message: "无法上传文件",
			Err:     nil,
		}
	}

	webhookData, err := json.Marshal(ReqWebhookMsg{
		MsgType: msgtypefile,
		File: &WebhookFile{
			MediaID: fileID,
		},
	})
	if err != nil {
		return wxrobotID, fileID, &SendError{
			Code:    -1,
			Message: "编码请求结构体为json错误",
			Err:     err,
		}
	}

	resp, err := http.Post(flagparser.Webhook, "application/json", bytes.NewBuffer(webhookData))
	if err != nil {
		return wxrobotID, fileID, &SendError{
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
		return wxrobotID, fileID, &SendError{
			Code:    -3,
			Message: "读取Body错误",
			Err:     err,
		}
	}

	var respWebhook RespWebhookMsg
	err = json.Unmarshal(respData, &respWebhook)
	if err != nil {
		return wxrobotID, fileID, &SendError{
			Code:    -4,
			Message: "将body解析成json错误",
			Err:     err,
		}
	}

	if respWebhook.ErrCode != 0 {
		return wxrobotID, fileID, &SendError{
			Code:    -5,
			Message: fmt.Sprintf("企业微信报告错误 (错误码: %d): %s", respWebhook.ErrCode, respWebhook.ErrMsg),
			Err:     nil,
		}
	}

	return wxrobotID, fileID, nil
}

func uploadMsgFile(msg string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create the form file part.
	part, err := writer.CreateFormFile("media", "message.txt")
	if err != nil {
		return "", &SendError{
			Code:    -3,
			Message: "创建FormData文件字段错误",
			Err:     err,
		}
	}

	// Copy the file content to the multipart writer.
	_, err = part.Write([]byte(msg))
	if err != nil {
		return "", &SendError{
			Code:    -3,
			Message: "写入FormData文件数据错误",
			Err:     err,
		}
	}

	// Close the multipart writer to finalize the body.
	err = writer.Close()
	if err != nil {
		return "", &SendError{
			Code:    -3,
			Message: "关闭FormData错误",
			Err:     err,
		}
	}

	// Prepare the request.
	req, err := http.NewRequest("POST", fmt.Sprintf("%s?key=%s&type=%s", upload_media, GetWebHookKey(), upload_file), body)
	if err != nil {
		return "", &SendError{
			Code:    -3,
			Message: "生成HTTP-POST请求错误",
			Err:     err,
		}
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", &SendError{
			Code:    -3,
			Message: "HTTP请求错误",
			Err:     err,
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &SendError{
			Code:    -3,
			Message: "读取Body错误",
			Err:     err,
		}
	}

	var respWebhook RespWebhookFile
	err = json.Unmarshal(respData, &respWebhook)
	if err != nil {
		return "", &SendError{
			Code:    -4,
			Message: "将body解析成json错误",
			Err:     err,
		}
	}

	if respWebhook.ErrCode != 0 {
		return "", &SendError{
			Code:    -5,
			Message: fmt.Sprintf("企业微信报告错误 (错误码: %d): %s", respWebhook.ErrCode, respWebhook.ErrMsg),
			Err:     nil,
		}
	}

	if respWebhook.Type != upload_file {
		return "", &SendError{
			Code:    -5,
			Message: fmt.Sprintf("文件上传类型有误，期望 %s ，实际 %s", upload_file, respWebhook.Type),
			Err:     nil,
		}
	}

	return respWebhook.MediaID, nil
}

func getWxRobotID(webhook string, msg string, t time.Time) string {
	text := fmt.Sprintf("WXROBOT-%s\n%s\n%d", webhook, msg, t.Unix())
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GetWebHookKey() string {
	onceGetKey.Do(func() {
		u, err := url.Parse(flagparser.Webhook)
		if err == nil {
			key = u.Query().Get("key")
		} else {
			key = ""
		}
	})
	return key
}
