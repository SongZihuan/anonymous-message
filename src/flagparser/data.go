package flagparser

var Debug bool = false
var Origin string = ""
var HttpAddress string = ":3352"
var RedisAddress string = "127.0.0.0.1:6379"
var RedisPassword string = ""
var RedisDB int = 0
var SMTPAddress string = "smtp.qiye.aliyun.com:465"
var SMTPUser string = ""
var SMTPPassword string = ""
var SMTPRecipient string = ""
var SMTPSender string = ""

var IMAPAddress string = "imap.qiye.aliyun.com:993"
var IMAPUser string = ""
var IMAPPassword string = ""
var IMAPRecipient string = ""
var IMAPMailBox string = "电子信箱"

var SQLitePath = ""
var SQLiteActiveClose = false

var Webhook string = ""

var _TimeZoom string = "Local"

var ShowOption = false
var DryRun = false
var Version = false
var License = false
var Report = false
