package flagparser

import resource "github.com/SongZihuan/anonymous-message"

var Debug bool = false

var Origin string = ""
var HttpAddress string = ":3352"
var WebURL string = "（暂无）"
var Name string = resource.Name

var SMTPAddress string = "smtp.qiye.aliyun.com:465"
var SMTPUser string = ""
var SMTPPassword string = ""
var IMAPAddress string = "imap.qiye.aliyun.com:993"
var IMAPUser string = ""
var IMAPPassword string = ""
var RecipientList string = ""
var NoticeList string = ""
var MailBox string = "电子信箱"

var SQLitePath = ""
var SQLiteActiveClose = false

var Webhook string = ""

var _TimeZone string = "Local"

var NotProxyProto bool = false

var ShowOption = false
var DryRun = false
var Version = false
var License = false
var Report = false
