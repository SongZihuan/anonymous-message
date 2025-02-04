package flagparser

import (
	"flag"
	"fmt"
	"os"
)

func initFlag() (err error) {
	defer func() {
		e := recover()
		if e != nil {
			err = fmt.Errorf("%v", e)
			return
		}
	}()

	flag.CommandLine.SetOutput(os.Stdout)

	flag.BoolVar(&Debug, "debug", Debug, "debug mode")

	flag.StringVar(&HttpAddress, "a", HttpAddress, "http server listen address")
	flag.StringVar(&HttpAddress, "address", HttpAddress, "http server listen address")
	flag.StringVar(&HttpAddress, "http-address", HttpAddress, "http server listen address")

	flag.StringVar(&Webhook, "w", Webhook, "wechat business robot webhook")
	flag.StringVar(&Webhook, "web-hook", Webhook, "wechat business robot webhook")
	flag.StringVar(&Webhook, "webhook", Webhook, "wechat business robot webhook")

	flag.StringVar(&RedisAddress, "redis-address", Webhook, "redis address")
	flag.StringVar(&RedisPassword, "redis-password", Webhook, "redis password")
	flag.IntVar(&RedisDB, "redis-db", RedisDB, "redis db")

	flag.StringVar(&SMTPAddress, "smtp-address", SMTPAddress, "smtp service address, example: smtp.qiye.aliyun.com:465")
	flag.StringVar(&SMTPUser, "smtp-user", SMTPUser, "smtp user name")
	flag.StringVar(&SMTPPassword, "smtp-password", SMTPPassword, "smtp password")
	flag.StringVar(&SMTPRecipient, "smtp-recipient", SMTPRecipient, "recipients, comma separated")

	flag.StringVar(&SQLitePath, "sqlite-path", SQLitePath, "sqlite path")
	flag.BoolVar(&SQLiteActiveClose, "sqlite-active-close", SQLiteActiveClose, "sqlite uses active shutdown. note: usually it does not need to be enabled.")

	flag.StringVar(&Origin, "origin", Origin, "cors allow origin")

	flag.BoolVar(&DryRun, "dry-run", DryRun, "only parser the options")

	flag.BoolVar(&Version, "version", Version, "show the version")
	flag.BoolVar(&Version, "v", Version, "show the version")

	flag.BoolVar(&License, "license", License, "show the license")
	flag.BoolVar(&License, "l", License, "show the license")

	flag.BoolVar(&Report, "report", Report, "show the report")
	flag.BoolVar(&Report, "r", Report, "show the report")

	flag.BoolVar(&ShowOption, "show-option", ShowOption, "show the option")
	flag.BoolVar(&ShowOption, "s", ShowOption, "show the option")

	flag.Parse()

	return nil
}
