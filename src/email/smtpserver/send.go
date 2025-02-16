package smtpserver

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	resource "github.com/SongZihuan/anonymous-message"
	"github.com/SongZihuan/anonymous-message/src/database"
	"github.com/SongZihuan/anonymous-message/src/emailtemplate"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"github.com/SongZihuan/anonymous-message/src/reqrate"
	"github.com/SongZihuan/anonymous-message/src/utils"
	"gopkg.in/gomail.v2"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// RateExp: 计算周期
// RateMaxCount: 周期内最大发件数量
const (
	RateExp      time.Duration = 10 * time.Minute
	RateMaxCount int64         = 1
)

const (
	SMTPSendTypeToSelf = "ToSelf"
	SMTPSendTypeError  = "Error"
	SMTPSendTypeThank  = "Thank"
)

var ErrRateLimit = fmt.Errorf("rate limit")

var toAddr []*mail.Address
var allowSender map[string]bool

func InitSmtp() error {
	if flagparser.SMTPAddress == "" || flagparser.SMTPUser == "" {
		return nil
	}

	recipientList := strings.Split(flagparser.SMTPRecipient, ",")
	toAddr = make([]*mail.Address, 0, len(recipientList))

	for _, rec := range recipientList {
		rec = strings.TrimSpace(rec)

		addr, err := mail.ParseAddress(rec)
		if err != nil {
			fmt.Printf("%s parser failled, ignore\n", rec)
			continue
		}

		if !utils.IsValidEmail(addr.Address) {
			fmt.Printf("%s is not a valid email, ignore\n", addr.Address)
			continue
		}

		toAddr = append(toAddr, addr)
	}

	if len(toAddr) == 0 {
		return fmt.Errorf("not any valid email address to be self recipient")
	}

	senderList := strings.Split(flagparser.SMTPSender, ",")
	allowSender = make(map[string]bool, len(recipientList)+1)

	allowSender[flagparser.SMTPUser] = true

	for _, rec := range senderList {
		rec = strings.TrimSpace(rec)

		addr, err := mail.ParseAddress(rec)
		if err != nil {
			fmt.Printf("%s parser failled, ignore\n", rec)
			continue
		}

		if !utils.IsValidEmail(addr.Address) {
			fmt.Printf("%s is not a valid email, ignore\n", addr.Address)
			continue
		}

		allowSender[addr.Address] = true
	}

	return nil
}

func SendToSelf(subject string, msg string, t time.Time) (string, error) {
	if flagparser.SMTPAddress == "" || flagparser.SMTPUser == "" {
		return "", notSMTPUser
	}

	subject = fmt.Sprintf("【%s 消息提醒】 %s", resource.Name, subject)

	smtpID, err := sendTo(subject, msg, nil, nil, toAddr, "", t)
	if err != nil {
		return "", err
	}

	_ = SMTPSendTypeToSelf // 假装使用一下这个常量

	return smtpID, nil
}

func SendThankMsg(subject string, messageID string, fromAddr *mail.Address, toAddr *mail.Address, replyAddr *mail.Address) (string, error) {
	if flagparser.SMTPAddress == "" || flagparser.SMTPUser == "" {
		return "", notSMTPUser
	}

	if replyAddr.Address == "" || !utils.IsValidEmail(replyAddr.Address) {
		return "", fmt.Errorf("invalid reply address")
	}

	if reqrate.CheckSMTPSendAddressRate(SMTPSendTypeThank, replyAddr, RateExp) > RateMaxCount {
		return "", ErrRateLimit
	}

	now := time.Now()

	fromName, err := utils.FormatEmailAddressToHumanStringJustNameSafe(fromAddr)
	if err != nil {
		fromName = fromAddr.Address
	}

	toNameAndAddr, err := utils.FormatEmailAddressToHumanStringSafe(toAddr)
	if err != nil {
		toAddr.Name = resource.Name
		toNameAndAddr = toAddr.String()
	}

	sender := flagparser.SMTPUser
	if yes, ok := allowSender[toAddr.Address]; ok && yes {
		sender = ""
	}

	data := &emailtemplate.ImapThankEmailModel{
		SenderAddr:    sender,
		FromName:      fromName,
		ToNameAndAddr: toNameAndAddr,
		ReplyAddr:     replyAddr.Address,
		SystemName:    resource.Name,
		Date:          now.In(flagparser.TimeZone()).Format("2006-01-02 15:04:05"),
		DateLocation:  flagparser.TimeZone().String(),
		DateUTC:       now.In(time.UTC).Format("2006-01-02 15:04:05"),
	}

	var tplResult bytes.Buffer
	err = emailtemplate.ImapThankEmail.Execute(&tplResult, data)
	if err != nil {
		return "", err
	}
	msg := tplResult.String()

	if messageID != "" && !strings.HasPrefix(strings.ToLower(subject), "re:") {
		subject = "Re: " + subject
	}

	from := toAddr
	to := []*mail.Address{replyAddr}

	smtpID, err := sendTo(subject, msg, from, nil, to, messageID, now)
	if err != nil {
		return "", err
	}

	return smtpID, nil
}

func SendErrorMsg(subject string, messageID string, fromAddr *mail.Address, toAddr *mail.Address, replyAddr *mail.Address, errorMsg string) (string, error) {
	if flagparser.SMTPAddress == "" || flagparser.SMTPUser == "" {
		return "", notSMTPUser
	}

	if errorMsg == "" {
		return "", fmt.Errorf("error msg is empty")
	}

	if replyAddr.Address == "" || !utils.IsValidEmail(replyAddr.Address) {
		return "", fmt.Errorf("invalid reply address")
	}

	if reqrate.CheckSMTPSendAddressRate(SMTPSendTypeError, replyAddr, RateExp) > RateMaxCount {
		return "", ErrRateLimit
	}

	now := time.Now()

	fromName, err := utils.FormatEmailAddressToHumanStringJustNameSafe(fromAddr)
	if err != nil {
		fromName = fromAddr.Address
	}

	toNameAndAddr, err := utils.FormatEmailAddressToHumanStringSafe(toAddr)
	if err != nil {
		toAddr.Name = resource.Name
		toNameAndAddr = toAddr.String()
	}

	sender := flagparser.SMTPUser
	if yes, ok := allowSender[toAddr.Address]; ok && yes {
		sender = ""
	}

	data := &emailtemplate.ImapErrorEmailModel{
		SenderAddr:    sender,
		FromName:      fromName,
		ToNameAndAddr: toNameAndAddr,
		ErrorMsg:      errorMsg,
		ReplyAddr:     replyAddr.Address,
		SystemName:    resource.Name,
		Date:          now.In(flagparser.TimeZone()).Format("2006-01-02 15:04:05"),
		DateLocation:  flagparser.TimeZone().String(),
		DateUTC:       now.In(time.UTC).Format("2006-01-02 15:04:05"),
	}

	var tplResult bytes.Buffer
	err = emailtemplate.ImapErrorEmail.Execute(&tplResult, data)
	if err != nil {
		return "", err
	}
	msg := tplResult.String()

	if messageID != "" && !strings.HasPrefix(strings.ToLower(subject), "re:") {
		subject = "Re: " + subject
	}

	from := toAddr
	to := []*mail.Address{replyAddr}

	smtpID, err := sendTo(subject, msg, from, nil, to, messageID, now)
	if err != nil {
		return "", err
	}

	return smtpID, nil
}

var notSMTPUser = fmt.Errorf("not smtp user")

func sendTo(subject string, msg string, fromAddr *mail.Address, replyToAddr *mail.Address, toAddr []*mail.Address, messageID string, t time.Time) (smtpID string, err error) {
	if flagparser.SMTPAddress == "" || flagparser.SMTPUser == "" {
		return smtpID, notSMTPUser
	}

	defer func() {
		if smtpID != "" {
			_ = database.UpdateSMTPRecord(smtpID, err)
		}
	}()

	defer func() {
		r := recover()
		if r != nil && err == nil {
			if _err, ok := r.(error); ok {
				err = _err
			} else {
				err = fmt.Errorf("panic: %v", r)
			}
		}
	}()

	sender := flagparser.SMTPUser

	if fromAddr == nil {
		fromAddr = &mail.Address{
			Name:    resource.Name,
			Address: flagparser.SMTPUser,
		}
	} else if yes, ok := allowSender[fromAddr.Address]; ok && yes {
		sender = fromAddr.Address
	}

	if replyToAddr == nil {
		replyToAddr = &mail.Address{
			Name:    fromAddr.Name,
			Address: fromAddr.Address,
		}
	}

	smtpID = getSMTPMailID(subject, msg, fromAddr, replyToAddr, toAddr, messageID, t)

	err = database.SaveSMTPRecord(smtpID, sender, subject, msg, fromAddr, toAddr, messageID, t)
	if err != nil {
		return "", err
	}

	const missingPort = "missing port in address"
	host, port, err := net.SplitHostPort(flagparser.SMTPAddress)
	var addrErr *net.AddrError
	if errors.As(err, &addrErr) {
		if addrErr.Err == missingPort {
			host = flagparser.SMTPAddress
			port = "25"
		} else {
			return smtpID, err
		}
	} else if err != nil {
		return smtpID, err
	}

	tlsconfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return smtpID, err
	}
	defer func() {
		_ = conn.Close()
	}()

	isSecureConn := false
	_conn := tls.Client(conn, tlsconfig)
	err = _conn.Handshake()
	if err == nil {
		conn = _conn
		isSecureConn = true
	}

	smtpClient, err := smtp.NewClient(conn, host)
	if err != nil {
		return smtpID, fmt.Errorf("new smtp client: %v", err)
	}
	defer func() {
		_ = smtpClient.Quit()
		smtpClient = nil
	}()

	hostname, err := os.Hostname()
	if err != nil {
		return smtpID, err
	}

	if err = smtpClient.Hello(hostname); err != nil {
		return smtpID, fmt.Errorf("hello: %v", err)
	}

	// If not using SMTPS, always use STARTTLS if available
	hasStartTLS, _ := smtpClient.Extension("STARTTLS")
	if !isSecureConn && hasStartTLS {
		if err = smtpClient.StartTLS(tlsconfig); err != nil {
			return smtpID, fmt.Errorf("start tls: %v", err)
		}
	}

	canAuth, options := smtpClient.Extension("AUTH")
	if canAuth {
		var auth smtp.Auth
		if strings.Contains(options, "CRAM-MD5") {
			auth = smtp.CRAMMD5Auth(sender, flagparser.SMTPPassword)
		} else if strings.Contains(options, "PLAIN") {
			auth = smtp.PlainAuth("", sender, flagparser.SMTPPassword, host)
		} else if strings.Contains(options, "LOGIN") {
			auth = LoginAuth(sender, flagparser.SMTPPassword)
		}

		if auth != nil {
			if err = smtpClient.Auth(auth); err != nil {
				return smtpID, fmt.Errorf("auth: %v", err)
			}
		}
	}

	err = smtpClient.Mail(sender)
	if err != nil {
		return smtpID, fmt.Errorf("mail: %v", err)
	}

	recList := make([]string, 0, len(toAddr))

	for _, addr := range toAddr {
		if addr.Address == "" || !utils.IsValidEmail(addr.Address) {
			fmt.Printf("%s is not a valid email, ignore\n", addr.Address)
			continue
		}

		err = smtpClient.Rcpt(addr.Address)
		if err != nil {
			fmt.Printf("%s set rcpt error: %s, ignore\n", addr.String(), err.Error())
			continue
		}

		recList = append(recList, addr.String())
	}

	if len(recList) == 0 {
		return smtpID, fmt.Errorf("no any valid recipient")
	}

	if fromAddr.Address == "" {
		fromAddr.Address = flagparser.SMTPUser
	}

	gomsg := gomail.NewMessage()
	gomsg.SetHeader("From", fromAddr.String())
	gomsg.SetHeader("To", recList...)
	gomsg.SetHeader("Reply-To", replyToAddr.String())
	gomsg.SetHeader("Subject", subject)
	gomsg.SetDateHeader("Date", t)
	if messageID != "" {
		gomsg.SetHeader("In-Reply-To", messageID)
		gomsg.SetHeader("References", messageID)
	}
	gomsg.SetBody("text/plain", msg)

	w, err := smtpClient.Data()
	if err != nil {
		return smtpID, fmt.Errorf("data: %v", err)
	}

	if _, err = gomsg.WriteTo(w); err != nil {
		return smtpID, fmt.Errorf("write to: %v", err)
	}

	err = w.Close()
	if err != nil {
		return smtpID, fmt.Errorf("close: %v", err)
	}

	return smtpID, nil
}

func getSMTPMailID(subject string, msg string, fromAddr *mail.Address, replyToAddr *mail.Address, toAddr []*mail.Address, messageID string, t time.Time) string {
	rec := make([]string, 0, len(toAddr))
	for _, to := range toAddr {
		rec = append(rec, to.String())
	}

	text := fmt.Sprintf("SMTP-%s\n%s\n%s\n%s\n%s\n%s\n%d", subject, msg, fromAddr.String(), replyToAddr.String(), strings.Join(rec, ";"), messageID, t.Unix())
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

type loginAuth struct {
	username, password string
}

func (*loginAuth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, fmt.Errorf("unknwon fromServer: %s", string(fromServer))
		}
	}
	return nil, nil
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

type Message struct {
	Info string // Message information for log purpose.
	*gomail.Message
	confirmChan chan struct{}
}
