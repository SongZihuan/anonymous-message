package smtpserver

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/database"
	"github.com/SongZihuan/anonymous-message/src/emailserver/emailaddress"
	"github.com/SongZihuan/anonymous-message/src/emailserver/tpl"
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

var ErrRateLimit = fmt.Errorf("rate limit")

var ready = false

func InitSmtp() (err error) {
	defer func() {
		if err == nil {
			ready = true
		}
	}()

	if flagparser.SMTPAddress == "" || flagparser.SMTPUser == "" {
		return fmt.Errorf("smtp not ready")
	}

	return nil
}

func SendToSelf(subject string, msg string, t time.Time) (string, error) {
	if !ready {
		return "", fmt.Errorf("smtp not ready")
	}

	subject = fmt.Sprintf("【%s 消息提醒】 %s", flagparser.Name, subject)

	smtpID, err := sendTo(subject, msg, emailaddress.DefaultRecipientAddress, emailaddress.DefaultRecipientAddress, emailaddress.DefaultRecipientAddress, emailaddress.NoticeAddressList, "", t)
	if err != nil {
		return "", err
	}

	return smtpID, nil
}

// SendThankMsg myAddr 代表我方（From） userAddr 代表对方（To） 由我方发往对方
func SendThankMsg(subject string, messageID string, myAddr *mail.Address, userAddr *mail.Address) (string, error) {
	if !ready {
		return "", fmt.Errorf("smtp not ready")
	}

	if !reqrate.CheckSMTPSendAddressRate(reqrate.SMTPSendTypeError, userAddr) {
		return "", ErrRateLimit
	}

	now := time.Now()
	myAddr.Name = flagparser.Name

	data := &tpl.ImapThankEmailModel{
		UserAddr:     userAddr.Address,
		UserName:     utils.FormatEmailAddressToHumanStringJustName(userAddr),
		MyNameAddr:   utils.FormatEmailAddressToHumanStringMustSafe(myAddr),
		MyAddr:       myAddr.Address,
		MyName:       flagparser.Name,
		Date:         now.In(flagparser.TimeZone()).Format("2006-01-02 15:04:05"),
		DateLocation: flagparser.TimeZone().String(),
		DateUTC:      now.In(time.UTC).Format("2006-01-02 15:04:05"),
		WebURL:       flagparser.WebURL,
	}

	var tplResult bytes.Buffer
	err := tpl.ImapThankEmail.Execute(&tplResult, data)
	if err != nil {
		return "", err
	}
	msg := tplResult.String()

	if messageID != "" && !strings.HasPrefix(strings.ToLower(subject), "re:") {
		subject = "Re: " + subject
	}

	smtpID, err := sendTo(subject, msg, myAddr, myAddr, myAddr, []*mail.Address{userAddr}, messageID, now)
	if err != nil {
		return smtpID, err
	}

	return smtpID, nil
}

// SendErrorMsg myAddr 代表我方（From） userAddr 代表对方（To） 由我方发往对方
func SendErrorMsg(subject string, messageID string, myAddr *mail.Address, userAddr *mail.Address, errorMsg string) (string, error) {
	if !ready {
		return "", fmt.Errorf("smtp not ready")
	}

	if errorMsg == "" {
		return "", fmt.Errorf("error msg is empty")
	}

	if !reqrate.CheckSMTPSendAddressRate(reqrate.SMTPSendTypeError, userAddr) {
		return "", ErrRateLimit
	}

	now := time.Now()
	myAddr.Name = flagparser.Name

	data := &tpl.ImapErrorEmailModel{
		UserAddr:     userAddr.Address,
		UserName:     utils.FormatEmailAddressToHumanStringJustName(userAddr),
		MyNameAddr:   utils.FormatEmailAddressToHumanStringMustSafe(myAddr),
		MyAddr:       myAddr.Address,
		MyName:       flagparser.Name,
		ErrorMsg:     errorMsg,
		Date:         now.In(flagparser.TimeZone()).Format("2006-01-02 15:04:05"),
		DateLocation: flagparser.TimeZone().String(),
		DateUTC:      now.In(time.UTC).Format("2006-01-02 15:04:05"),
		WebURL:       flagparser.WebURL,
	}

	var tplResult bytes.Buffer
	err := tpl.ImapErrorEmail.Execute(&tplResult, data)
	if err != nil {
		return "", err
	}
	msg := tplResult.String()

	if messageID != "" && !strings.HasPrefix(strings.ToLower(subject), "re:") {
		subject = "Re: " + subject
	}

	smtpID, err := sendTo(subject, msg, myAddr, myAddr, myAddr, []*mail.Address{userAddr}, messageID, now)
	if err != nil {
		return smtpID, err
	}

	return smtpID, nil
}

var notSMTPUser = fmt.Errorf("not smtp user")

func sendTo(subject string, msg string, senderAddr *mail.Address, fromAddr *mail.Address, replyToAddr *mail.Address, toAddr []*mail.Address, messageID string, t time.Time) (smtpID string, err error) {
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

	if senderAddr == nil {
		senderAddr = emailaddress.DefaultRecipientAddress
	}

	if fromAddr == nil {
		fromAddr = &mail.Address{
			Name:    senderAddr.Name,
			Address: senderAddr.Address,
		}
	}

	if replyToAddr == nil {
		replyToAddr = &mail.Address{
			Name:    fromAddr.Name,
			Address: fromAddr.Address,
		}
	}

	smtpID = getSMTPMailID(subject, msg, senderAddr, fromAddr, replyToAddr, toAddr, messageID, t)

	err = database.SaveSMTPRecord(smtpID, subject, msg, senderAddr, fromAddr, toAddr, messageID, t)
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
			auth = smtp.CRAMMD5Auth(senderAddr.Address, flagparser.SMTPPassword)
		} else if strings.Contains(options, "PLAIN") {
			auth = smtp.PlainAuth("", senderAddr.Address, flagparser.SMTPPassword, host)
		} else if strings.Contains(options, "LOGIN") {
			auth = LoginAuth(senderAddr.Address, flagparser.SMTPPassword)
		}

		if auth != nil {
			if err = smtpClient.Auth(auth); err != nil {
				return smtpID, fmt.Errorf("auth: %v", err)
			}
		}
	}

	err = smtpClient.Mail(senderAddr.Address)
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

func getSMTPMailID(subject string, msg string, senderAddr *mail.Address, fromAddr *mail.Address, replyToAddr *mail.Address, toAddr []*mail.Address, messageID string, t time.Time) string {
	rec := make([]string, 0, len(toAddr))
	for _, to := range toAddr {
		rec = append(rec, to.String())
	}

	text := fmt.Sprintf("SMTP-%s\n%s\n%s\n%s\n%s\n%s\n%s\n%d", subject, msg, senderAddr.String(), fromAddr.String(), replyToAddr.String(), strings.Join(rec, ";"), messageID, t.Unix())
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
