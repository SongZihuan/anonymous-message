package email

import (
	"crypto/tls"
	"errors"
	"fmt"
	resource "github.com/SongZihuan/anonymous-message"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"github.com/SongZihuan/anonymous-message/src/utils"
	"gopkg.in/gomail.v2"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"
)

func Send(title string, msg string, t time.Time) error {
	if flagparser.SMTPAddress == "" || flagparser.SMTPUser == "" || flagparser.SMTPRecipient == "" {
		return nil
	}

	sender := flagparser.SMTPUser

	const missingPort = "missing port in address"
	host, port, err := net.SplitHostPort(flagparser.SMTPAddress)
	var addrErr *net.AddrError
	if errors.As(err, &addrErr) {
		if addrErr.Err == missingPort {
			host = flagparser.SMTPAddress
			port = "25"
		} else {
			return err
		}
	} else if err != nil {
		return err
	}

	tlsconfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return err
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

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("new smtp client: %v", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	if err = client.Hello(hostname); err != nil {
		return fmt.Errorf("hello: %v", err)
	}

	// If not using SMTPS, always use STARTTLS if available
	hasStartTLS, _ := client.Extension("STARTTLS")
	if !isSecureConn && hasStartTLS {
		if err = client.StartTLS(tlsconfig); err != nil {
			return fmt.Errorf("start tls: %v", err)
		}
	}

	canAuth, options := client.Extension("AUTH")
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
			if err = client.Auth(auth); err != nil {
				return fmt.Errorf("auth: %v", err)
			}
		}
	}

	err = client.Mail(sender)
	if err != nil {
		return fmt.Errorf("mail: %v", err)
	}

	var recipientList = strings.Split(flagparser.SMTPRecipient, ",")
	var recList = make([]string, 0, len(recipientList))
	for _, rec := range recipientList {
		rec = strings.TrimSpace(rec)
		if !utils.IsValidEmail(rec) {
			fmt.Printf("%s is not a valid email, ignore\n", rec)
		}

		err = client.Rcpt(rec)
		if err != nil {
			fmt.Printf("%s set rcpt error: %s, ignore\n", rec, err.Error())
		}

		recList = append(recList, rec)
	}

	if len(recList) == 0 {
		return fmt.Errorf("no any valid recipient")
	}

	fromAddr := mail.Address{
		Name:    resource.Name,
		Address: sender,
	}

	gomsg := gomail.NewMessage()
	gomsg.SetHeader("From", fromAddr.String())
	gomsg.SetHeader("To", recList...)
	gomsg.SetHeader("Subject", fmt.Sprintf("【%s】%s", resource.Name, title))
	gomsg.SetDateHeader("Date", t)
	gomsg.SetBody("text/plain", msg)

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %v", err)
	}

	if _, err = gomsg.WriteTo(w); err != nil {
		return fmt.Errorf("write to: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("close: %v", err)
	}

	return client.Quit()
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
