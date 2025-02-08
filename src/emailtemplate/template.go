package emailtemplate

import (
	_ "embed"
	"text/template"
)

//go:embed imap_error_email.txtmpl
var imapErrorEmail string

//go:embed imap_thank_email.txtmpl
var imapThankEmail string

var ImapErrorEmail *template.Template
var ImapThankEmail *template.Template

type ImapErrorEmailModel struct {
	FromName      string
	ToNameAndAddr string
	ErrorMsg      string
	ReplyAddr     string
	SystemName    string
	Date          string
	DateLocation  string
	DateUTC       string
}

type ImapThankEmailModel struct {
	FromName      string
	ToNameAndAddr string
	ReplyAddr     string
	SystemName    string
	Date          string
	DateLocation  string
	DateUTC       string
}

func init() {
	var err error

	ImapErrorEmail, err = template.New("ImapErrorEmail").Parse(imapErrorEmail)
	if err != nil {
		panic(err)
	}

	ImapThankEmail, err = template.New("ImapThankEmail").Parse(imapThankEmail)
	if err != nil {
		panic(err)
	}
}
