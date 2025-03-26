package tpl

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
	UserAddr     string
	UserName     string
	MyAddr       string
	MyName       string
	MyNameAddr   string
	ErrorMsg     string
	Date         string
	DateLocation string
	DateUTC      string
	WebURL       string
}

type ImapThankEmailModel struct {
	UserAddr     string
	UserName     string
	MyAddr       string
	MyName       string
	MyNameAddr   string
	Date         string
	DateLocation string
	DateUTC      string
	WebURL       string
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
