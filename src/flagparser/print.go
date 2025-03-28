package flagparser

import (
	"flag"
	"fmt"
	resource "github.com/SongZihuan/anonymous-message"
	"github.com/SongZihuan/anonymous-message/src/utils"
)

func PrintLicense() (int, error) {
	title := utils.FormatTextToWidth(fmt.Sprintf("License of %s:", utils.GetArgs0Name()), utils.NormalConsoleWidth)
	license := utils.FormatTextToWidth(resource.License, utils.NormalConsoleWidth)
	return fmt.Fprintf(flag.CommandLine.Output(), "%s\n%s\n", title, license)
}

func PrintVersion() (int, error) {
	version := utils.FormatTextToWidth(fmt.Sprintf("Version of %s: %s", utils.GetArgs0Name(), resource.Version), utils.NormalConsoleWidth)
	return fmt.Fprintf(flag.CommandLine.Output(), "%s\n", version)
}

func PrintReport() (int, error) {
	// 不需要title
	report := utils.FormatTextToWidth(resource.Report, utils.NormalConsoleWidth)
	return fmt.Fprintf(flag.CommandLine.Output(), "%s\n", report)
}

func PrintLF() (int, error) {
	return fmt.Fprintf(flag.CommandLine.Output(), "\n")
}

func Print() {
	fmt.Println("Debug:", Debug)
	fmt.Println("Origin:", Origin)
	fmt.Println("HttpAddress:", HttpAddress)
	fmt.Println("WebURL:", WebURL)
	fmt.Println("Not Use Proxy Proto:", NotProxyProto)
	fmt.Println("Webhook:", Webhook)
	fmt.Println("SMTP Address:", SMTPAddress)
	fmt.Println("SMTP User Name:", SMTPUser)
	fmt.Println("SMTP Password:", SMTPPassword)
	fmt.Println("IMAP Address:", IMAPAddress)
	fmt.Println("IMAP User Name:", IMAPUser)
	fmt.Println("IMAP Password:", IMAPPassword)
	fmt.Println("SMTP Recipient:", NoticeList)
	fmt.Println("IMAP Recipient:", RecipientList)
	fmt.Println("IMAP MailBox:", MailBox)
	fmt.Println("SQLite Path:", SQLitePath)
	fmt.Println("SQLite Active Close:", SQLiteActiveClose)
	fmt.Println("Time Zone (use set) : ", _TimeZone)
	fmt.Println("Time Zone: ", TimeZone())
}
