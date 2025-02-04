package database

import (
	"database/sql"
	"gorm.io/gorm"
	"time"
)

// Model gorm.Model的仿写，明确了键名
type Model struct {
	ID        uint           `gorm:"column:id;primarykey"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;primarykey"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;primarykey"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

type Mail struct {
	Model
	MailID      string         `gorm:"column:mail_id;type:VARCHAR(100);not null;uniqueIndex;"`
	Name        string         `gorm:"column:name;type:VARCHAR(40);not null"`
	Content     string         `gorm:"column:content;type:VARCHAR(250);not null"`
	Refer       string         `gorm:"column:refer;type:VARCHAR(60);not null"`
	Origin      string         `gorm:"column:origin;type:VARCHAR(60);not null"`
	Host        string         `gorm:"column:host;type:VARCHAR(60);not null"`
	IP          string         `gorm:"column:ip;type:VARCHAR(50);not null"`
	Time        time.Time      `gorm:"column:time;not null"`
	Message     string         `gorm:"column:message;type:VARCHAR(4096);not null"`
	SendWxRobot bool           `gorm:"column:send_wx_robot;not null"`
	WxRobotErr  sql.NullString `orm:"column:wx_robot_err;type:VARCHAR(200);"`
	SendEmail   bool           `gorm:"column:send_email;not null"`
	EmailErr    sql.NullString `gorm:"column:email_err;type:VARCHAR(200);"`
	SystemName  string         `gorm:"column:system_name;type:VARCHAR(20);not null"`
	Version     string         `gorm:"column:version;type:VARCHAR(20);not null"`
}

func (*Mail) TableName() string {
	return "mail"
}
