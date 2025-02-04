package database

import (
	"database/sql"
	"errors"
	"fmt"
	resource "github.com/SongZihuan/anonymous-message"
	"github.com/SongZihuan/anonymous-message/src/utils"
	"gorm.io/gorm"
	"time"
)

func SaveMail(mailID string, name string, content string, refer string, origin string, host string, clientIP string, t time.Time, message string) error {
	if db == nil {
		return nil
	}

	if utils.GetMailID(name, content, refer, origin, host, t) != mailID {
		return fmt.Errorf("mail id check failed")
	}

	if len(origin) > 50 {
		origin = origin[:50]
	}

	if len(host) > 55 {
		host = host[:55]
	}

	version := resource.Version
	if len(version) > 15 {
		version = version[:15]
	}

	systemName := resource.Name
	if len(systemName) > 15 {
		systemName = systemName[:15]
	}

	mail := &Mail{
		MailID:      mailID,
		Name:        name,
		Content:     content,
		Refer:       refer,
		Origin:      origin,
		Host:        host,
		IP:          clientIP,
		Time:        t,
		Message:     message,
		SendWxRobot: false,
		SendEmail:   false,
		Version:     version,
		SystemName:  systemName,
	}

	err := db.Create(mail).Error
	if err != nil {
		return err
	}

	return nil
}

func UpdateWxRobotSendMsg(mailID string, sendErr error) error {
	if db == nil {
		return nil
	}

	var mail Mail
	err := db.Model(&Mail{}).Where("mail_id = ?", mailID).Order("time desc").First(&mail).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("mail not found")
	}

	if mail.SendWxRobot {
		return nil
	}

	mail.SendWxRobot = true
	if sendErr == nil {
		mail.WxRobotErr = sql.NullString{
			Valid: false,
		}
	} else {
		errMsg := sendErr.Error()

		if len(errMsg) > 190 {
			errMsg = errMsg[:190]
		}

		mail.WxRobotErr = sql.NullString{
			Valid:  true,
			String: errMsg,
		}
	}

	err = db.Save(&mail).Error
	if err != nil {
		return err
	}

	return nil
}

func UpdateEmailSendMsg(mailID string, sendErr error) error {
	if db == nil {
		return nil
	}

	var mail Mail
	err := db.Model(&Mail{}).Where("mail_id = ?", mailID).Order("time desc").First(&mail).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("mail not found")
	}

	if mail.SendEmail {
		return nil
	}

	mail.SendEmail = true
	if sendErr == nil {
		mail.EmailErr = sql.NullString{
			Valid: false,
		}
	} else {
		errMsg := sendErr.Error()

		if len(errMsg) > 190 {
			errMsg = errMsg[:190]
		}

		mail.EmailErr = sql.NullString{
			Valid:  true,
			String: errMsg,
		}
	}

	err = db.Save(&mail).Error
	if err != nil {
		return err
	}

	return nil
}
