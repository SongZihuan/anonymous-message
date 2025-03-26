package database

import (
	"database/sql"
	"gorm.io/gorm"
	"time"
)

// Model gorm.Model的仿写，明确了键名
type Model struct {
	ID        uint           `gorm:"column:id;primarykey"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

type MsgType string

const (
	MsgTypeWebsite MsgType = "WebsiteMessage - 网页留言"
	MsgTypeEmail   MsgType = "EmailMessage - 邮箱留言"
	MsgTypeSystem  MsgType = "SystemMessage - 系统留言"
)

type MailRecord struct {
	Model
	MailID   string  `gorm:"column:mail_id;type:VARCHAR(100);not null;uniqueIndex;"`
	MailType MsgType `gorm:"column:mail_type;type:VARCHAR(100);not null;"`
}

func (*MailRecord) TableName() string {
	return "all_mail"
}

type AMMail struct {
	Model
	MailID       string         `gorm:"column:mail_id;type:VARCHAR(100);not null;uniqueIndex;"`
	Name         string         `gorm:"column:name;type:VARCHAR(40);not null"`
	Email        string         `gorm:"column:email;type:VARCHAR(128);not null"`
	Content      string         `gorm:"column:content;type:TEXT;not null"`
	Refer        string         `gorm:"column:refer;type:VARCHAR(60);not null"`
	Origin       string         `gorm:"column:origin;type:VARCHAR(60);not null"`
	Host         string         `gorm:"column:host;type:VARCHAR(60);not null"`
	IP           string         `gorm:"column:ip;type:VARCHAR(50);not null"`
	Time         time.Time      `gorm:"column:time;not null"`
	WxRobotID    sql.NullString `gorm:"column:email_id;type:VARCHAR(100);"`
	EmailID      sql.NullString `gorm:"column:email_id;type:VARCHAR(100);"`
	ThankEmailID sql.NullString `gorm:"column:thank_email_id;type:VARCHAR(100);"`
	SystemName   string         `gorm:"column:system_name;type:VARCHAR(20);not null"`
	Version      string         `gorm:"column:version;type:VARCHAR(20);not null"`
}

func (*AMMail) TableName() string {
	return "am_mail"
}

type IMAPMail struct {
	Model
	MailID       string         `gorm:"column:mail_id;type:VARCHAR(100);not null;uniqueIndex;"`
	MessageID    string         `gorm:"column:message_id;type:VARCHAR(128);not null"`
	Sender       string         `gorm:"column:sender;type:VARCHAR(128);not null"`
	From         string         `gorm:"column:from;type:VARCHAR(128);not null"`
	To           string         `gorm:"column:to;type:VARCHAR(128);not null"`
	ReplyTo      string         `gorm:"column:reply_to;type:VARCHAR(128);not null"`
	Subject      string         `gorm:"column:subject;type:VARCHAR(128);not null"`
	Content      string         `gorm:"column:content;type:TEXT;not null"`
	SendTime     time.Time      `gorm:"column:send_time;not null"`
	Time         time.Time      `gorm:"column:time;not null"`
	WxRobotID    sql.NullString `gorm:"column:email_id;type:VARCHAR(100);"`
	EmailID      sql.NullString `gorm:"column:email_id;type:VARCHAR(100);"`
	ThankEmailID sql.NullString `gorm:"column:thank_email_id;type:VARCHAR(100);"`
	SystemName   string         `gorm:"column:system_name;type:VARCHAR(20);not null"`
	Version      string         `gorm:"column:version;type:VARCHAR(20);not null"`
}

func (*IMAPMail) TableName() string {
	return "imap_mail"
}

type SystemNotifyMail struct {
	Model
	MailID     string         `gorm:"column:mail_id;type:VARCHAR(100);not null;uniqueIndex;"`
	Subject    string         `gorm:"column:subject;type:VARCHAR(128);not null"`
	Content    string         `gorm:"column:content;type:VARCHAR(4096);not null"` // 限制长度在2048以内
	Time       time.Time      `gorm:"column:time;not null"`
	WxRobotID  sql.NullString `gorm:"column:email_id;type:VARCHAR(100);"`
	EmailID    sql.NullString `gorm:"column:email_id;type:VARCHAR(100);"`
	SystemName string         `gorm:"column:system_name;type:VARCHAR(20);not null"`
	Version    string         `gorm:"column:version;type:VARCHAR(20);not null"`
}

func (*SystemNotifyMail) TableName() string {
	return "sys_mail"
}

type WxRobotRecord struct {
	Model
	WxRobotID string `gorm:"column:wxrobot_id;type:VARCHAR(100);not null;uniqueIndex;"`
	Webhook   string `gorm:"column:webhook;type:VARCHAR(128);not null"`

	Content string         `gorm:"column:content;type:TEXT;not null"`
	Success bool           `gorm:"column:success;not null"`
	ErrMsg  sql.NullString `gorm:"column:err_msg;type:VARCHAR(200);"`

	FileContent sql.NullString `gorm:"column:file_time"`
	FileSuccess sql.NullBool   `gorm:"column:file_success"`
	FileErrMsg  sql.NullString `gorm:"column:file_err_msg;type:VARCHAR(200);"`
	FileID      sql.NullString `gorm:"column:file_id;type:VARCHAR(128)"`

	Time       time.Time `gorm:"column:time;not null"`
	SystemName string    `gorm:"column:system_name;type:VARCHAR(20);not null"`
	Version    string    `gorm:"column:version;type:VARCHAR(20);not null"`
}

func (*WxRobotRecord) TableName() string {
	return "wxrobot_record"
}

type SMTPRecord struct {
	Model
	SmtpID         string         `gorm:"column:smtp_id;type:VARCHAR(100);not null;uniqueIndex;"`
	Sender         string         `gorm:"column:sender;type:VARCHAR(128);not null"`
	From           string         `gorm:"column:from;type:VARCHAR(128);not null"`
	Subject        string         `gorm:"column:subject;type:VARCHAR(128);not null"`
	Content        string         `gorm:"column:content;type:TEXT;not null"`
	ReplyMessageID sql.NullString `gorm:"column:content;type:VARCHAR(1024)"`
	Time           time.Time      `gorm:"column:time;not null"`
	Success        bool           `gorm:"column:success;not null"`
	ErrMsg         sql.NullString `gorm:"column:err_msg;type:VARCHAR(200);"`
	SystemName     string         `gorm:"column:system_name;type:VARCHAR(20);not null"`
	Version        string         `gorm:"column:version;type:VARCHAR(20);not null"`
}

func (*SMTPRecord) TableName() string {
	return "smtp_record"
}

type SMTPRecipientRecord struct {
	Model
	SmtpID    string `gorm:"column:smtp_id;type:VARCHAR(100);not null;"`
	Recipient string `gorm:"column:recipient;type:VARCHAR(128);not null"`
}

func (*SMTPRecipientRecord) TableName() string {
	return "smtp_recipient_record"
}
