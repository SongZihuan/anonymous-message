package database

import (
	"database/sql"
	"errors"
	"fmt"
	resource "github.com/SongZihuan/anonymous-message"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"github.com/SongZihuan/anonymous-message/src/utils"
	"gorm.io/gorm"
	netmail "net/mail"
	"time"
)

var ErrNotFound = errors.New("record not found")

func SaveMailRecord(MailID string, MailType MsgType) error {
	if db == nil {
		return nil
	}

	record := MailRecord{
		MailID:   MailID,
		MailType: MailType,
	}

	err := db.Create(&record).Error
	if err != nil {
		return err
	}

	return nil
}

func SaveAMMail(mailID string, name string, email string, content string, refer string, origin string, host string, clientIP string, t time.Time) error {
	if db == nil {
		return nil
	}

	if utils.GetAMMailID(name, email, content, refer, origin, host, t) != mailID {
		return fmt.Errorf("mail id check failed")
	}

	mail := &AMMail{
		MailID:  mailID,
		Name:    name,
		Email:   email,
		Content: content,
		Refer:   refer,
		Origin:  origin,
		Host:    host,
		IP:      clientIP,
		Time:    t,
	}

	err := db.Create(mail).Error
	if err != nil {
		return err
	}

	err = SaveMailRecord(mailID, MsgTypeWebsite)
	if err != nil {
		return err
	}

	return nil
}

func SaveIMAPMail(mailID string, messageID string, sender string, from string, to string, replyTo string, subject string, content string, date time.Time, t time.Time) error {
	if db == nil {
		return nil
	}

	if utils.GetIMAPMailID(messageID, sender, from, to, replyTo, subject, content, date, t) != mailID {
		return fmt.Errorf("mail id check failed")
	}

	if len(messageID) > 120 {
		messageID = messageID[:120]
	}

	if len(sender) > 120 {
		sender = sender[:120]
	}

	if len(from) > 120 {
		from = from[:120]
	}

	if len(to) > 120 {
		to = to[:120]
	}

	if len(replyTo) > 120 {
		replyTo = replyTo[:120]
	}

	if len(subject) > 120 {
		subject = subject[:120]
	}

	mail := &IMAPMail{
		MailID:     mailID,
		MessageID:  messageID,
		Sender:     sender,
		From:       from,
		To:         to,
		ReplyTo:    replyTo,
		Subject:    subject,
		Content:    content,
		SendTime:   date,
		Time:       t,
		Version:    resource.Version,
		SystemName: flagparser.Name,
	}

	err := db.Create(mail).Error
	if err != nil {
		return err
	}

	err = SaveMailRecord(mailID, MsgTypeEmail)
	if err != nil {
		return err
	}

	return nil
}

func SaveSNMail(mailID string, subject string, content string, t time.Time) error {
	if db == nil {
		return nil
	}

	if utils.GetSNMailID(subject, content, t) != mailID {
		return fmt.Errorf("mail id check failed")
	}

	if len(subject) > 120 {
		if _subject, ok := utils.CompressAuto(subject, 120); ok {
			subject = _subject
		} else {
			subject = _subject[:120]
		}
	}

	if len(content) > 2040 {
		if _content, ok := utils.CompressAuto(content, 2040); ok {
			content = _content
		} else {
			content = _content[:2040]
		}
	}

	mail := &SystemNotifyMail{
		MailID:     mailID,
		Subject:    subject,
		Content:    content,
		Time:       t,
		Version:    resource.Version,
		SystemName: flagparser.Name,
	}

	err := db.Create(mail).Error
	if err != nil {
		return err
	}

	err = SaveMailRecord(mailID, MsgTypeSystem)
	if err != nil {
		return err
	}

	return nil
}

func UpdateAMWxRobotSendMsg(mailID string, wxrobotID string) error {
	if db == nil {
		return nil
	}

	var mail AMMail
	err := db.Model(&AMMail{}).Where("mail_id = ?", mailID).Order("time desc").First(&mail).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("mail not found")
	} else if err != nil {
		return err
	}

	mail.WxRobotID = sql.NullString{
		Valid:  wxrobotID != "",
		String: wxrobotID,
	}

	err = db.Save(&mail).Error
	if err != nil {
		return err
	}

	return nil
}

func UpdateIMAPWxRobotSendMsg(mailID string, wxrobotID string) error {
	if db == nil {
		return nil
	}

	var mail IMAPMail
	err := db.Model(&IMAPMail{}).Where("mail_id = ?", mailID).Order("time desc").First(&mail).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("mail not found")
	} else if err != nil {
		return err
	}

	mail.WxRobotID = sql.NullString{
		Valid:  wxrobotID != "",
		String: wxrobotID,
	}

	err = db.Save(&mail).Error
	if err != nil {
		return err
	}

	return nil
}

func UpdateSNWxRobotSendMsg(mailID string, wxrobotID string) error {
	if db == nil {
		return nil
	}

	var mail SystemNotifyMail
	err := db.Model(&SystemNotifyMail{}).Where("mail_id = ?", mailID).Order("time desc").First(&mail).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("mail not found")
	} else if err != nil {
		return err
	}

	mail.WxRobotID = sql.NullString{
		Valid:  wxrobotID != "",
		String: wxrobotID,
	}

	err = db.Save(&mail).Error
	if err != nil {
		return err
	}

	return nil
}

func UpdateAMEmailSendMsg(mailID string, smtpID string) error {
	if db == nil {
		return nil
	}

	var mail AMMail
	err := db.Model(&AMMail{}).Where("mail_id = ?", mailID).Order("time desc").First(&mail).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("mail not found")
	} else if err != nil {
		return err
	}

	mail.EmailID = sql.NullString{
		Valid:  smtpID != "",
		String: smtpID,
	}

	err = db.Save(&mail).Error
	if err != nil {
		return err
	}

	return nil
}

func UpdateIMAPEmailSendMsg(mailID string, smtpID string) error {
	if db == nil {
		return nil
	}

	var mail IMAPMail
	err := db.Model(&IMAPMail{}).Where("mail_id = ?", mailID).Order("time desc").First(&mail).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("mail not found")
	} else if err != nil {
		return err
	}

	mail.EmailID = sql.NullString{
		Valid:  smtpID != "",
		String: smtpID,
	}

	err = db.Save(&mail).Error
	if err != nil {
		return err
	}

	return nil
}

func UpdateSNEmailSendMsg(mailID string, smtpID string) error {
	if db == nil {
		return nil
	}

	var mail SystemNotifyMail
	err := db.Model(&SystemNotifyMail{}).Where("mail_id = ?", mailID).Order("time desc").First(&mail).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("mail not found")
	} else if err != nil {
		return err
	}

	mail.EmailID = sql.NullString{
		Valid:  smtpID != "",
		String: smtpID,
	}

	err = db.Save(&mail).Error
	if err != nil {
		return err
	}

	return nil
}

func UpdateAMThankEmailSendMsg(mailID string, smtpID string) error {
	if db == nil {
		return nil
	}

	var mail AMMail
	err := db.Model(&AMMail{}).Where("mail_id = ?", mailID).Order("time desc").First(&mail).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("mail not found")
	} else if err != nil {
		return err
	}

	mail.ThankEmailID = sql.NullString{
		Valid:  smtpID != "",
		String: smtpID,
	}

	err = db.Save(&mail).Error
	if err != nil {
		return err
	}

	return nil
}

func UpdateIMAPThankEmailSendMsg(mailID string, smtpID string) error {
	if db == nil {
		return nil
	}

	var mail IMAPMail
	err := db.Model(&IMAPMail{}).Where("mail_id = ?", mailID).Order("time desc").First(&mail).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("mail not found")
	} else if err != nil {
		return err
	}

	mail.ThankEmailID = sql.NullString{
		Valid:  smtpID != "",
		String: smtpID,
	}

	err = db.Save(&mail).Error
	if err != nil {
		return err
	}

	return nil
}

func FindIMAPMessageID(messageID string) (*IMAPMail, error) {
	var mail IMAPMail
	err := db.Model(&IMAPMail{}).Where("message_id = ?", messageID).Order("time desc").First(&mail).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &mail, nil
}

func SaveWxRobotRecord(wxrobotID string, webhook string, msg string, t time.Time) error {
	if db == nil {
		return nil
	}

	if len(msg) > 10240 {
		msg = msg[:10240]
	}

	record := WxRobotRecord{
		WxRobotID:  wxrobotID,
		Webhook:    webhook,
		Content:    msg,
		Time:       t,
		SystemName: flagparser.Name,
		Version:    resource.Version,
	}

	err := db.Create(&record).Error
	if err != nil {
		return err
	}

	return nil
}

func UpdateWxRobotRecord(wxrobotID string, sendErr error) error {
	if db == nil {
		return nil
	}

	var record WxRobotRecord
	err := db.Model(&WxRobotRecord{}).Where("wxrobot_id = ?", wxrobotID).Order("time desc").First(&record).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("record not found")
	} else if err != nil {
		return err
	}

	if sendErr == nil {
		record.Success = true
		record.ErrMsg = sql.NullString{
			Valid: false,
		}
	} else {
		errMsg := sendErr.Error()
		if len(errMsg) > 190 {
			errMsg = errMsg[:190]
		}

		record.Success = false
		record.ErrMsg = sql.NullString{
			Valid:  true,
			String: errMsg,
		}
	}

	err = db.Save(&record).Error
	if err != nil {
		return err
	}

	return nil
}

func SaveWxRobotFileRecord(wxrobotID string, msg string) error {
	if db == nil {
		return nil
	}

	var record WxRobotRecord
	err := db.Model(&WxRobotRecord{}).Where("wxrobot_id = ?", wxrobotID).Order("time desc").First(&record).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("record not found")
	} else if err != nil {
		return err
	}

	if len(msg) > 10240 {
		msg = msg[:10240]
	}

	record.FileContent = sql.NullString{
		Valid:  true,
		String: msg,
	}

	err = db.Save(&record).Error
	if err != nil {
		return err
	}

	return nil
}

func UpdateWxRobotFileRecord(wxrobotID string, fileID string, sendErr error) error {
	if db == nil {
		return nil
	}

	var record WxRobotRecord
	err := db.Model(&WxRobotRecord{}).Where("wxrobot_id = ?", wxrobotID).Order("time desc").First(&record).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("record not found")
	} else if err != nil {
		return err
	}

	record.FileID = sql.NullString{
		Valid:  fileID != "",
		String: fileID,
	}

	if sendErr == nil {
		record.FileSuccess = sql.NullBool{
			Valid: true,
			Bool:  true,
		}
		record.FileErrMsg = sql.NullString{
			Valid: false,
		}
	} else {
		errMsg := sendErr.Error()
		if len(errMsg) > 190 {
			errMsg = errMsg[:190]
		}

		record.FileSuccess = sql.NullBool{
			Valid: true,
			Bool:  false,
		}
		record.FileErrMsg = sql.NullString{
			Valid:  true,
			String: errMsg,
		}
	}

	err = db.Save(&record).Error
	if err != nil {
		return err
	}

	return nil
}

func SaveSMTPRecord(smtpID string, subject string, msg string, senderAddr *netmail.Address, fromAddr *netmail.Address, toAddr []*netmail.Address, messageID string, t time.Time) error {
	if db == nil {
		return nil
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		sender := senderAddr.String()

		if len(sender) > 100 {
			sender = sender[:100]
		}

		from := fromAddr.String()

		if len(from) > 100 {
			from = fromAddr.Address
		}

		if len(from) > 100 {
			from = from[:100]
		}

		if len(subject) > 100 {
			subject = subject[:100]
		}

		if len(messageID) > 1020 {
			messageID = messageID[:1020]
		}

		if len(msg) > 10240 {
			msg = msg[:10240]
		}

		record := SMTPRecord{
			SmtpID:  smtpID,
			Sender:  sender,
			From:    from,
			Subject: subject,
			Content: msg,
			ReplyMessageID: sql.NullString{
				Valid:  messageID != "",
				String: messageID,
			},
			Time:       t,
			SystemName: flagparser.Name,
			Version:    resource.Version,
		}

		err := tx.Create(&record).Error
		if err != nil {
			return err
		}

		for _, to := range toAddr {
			recipient := to.String()

			if len(recipient) > 100 {
				recipient = to.Address
			}

			if len(recipient) > 100 {
				recipient = recipient[:100]
			}

			recRecord := SMTPRecipientRecord{
				SmtpID:    smtpID,
				Recipient: recipient,
			}

			err := tx.Create(&recRecord).Error
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func UpdateSMTPRecord(smtpID string, sendErr error) error {
	if db == nil {
		return nil
	}

	var record SMTPRecord
	err := db.Model(&SMTPRecord{}).Where("smtp_id = ?", smtpID).Order("time desc").First(&record).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("record not found")
	} else if err != nil {
		return err
	}

	if sendErr == nil {
		record.Success = true
		record.ErrMsg = sql.NullString{
			Valid: false,
		}
	} else {
		errMsg := sendErr.Error()
		if len(errMsg) > 190 {
			errMsg = errMsg[:190]
		}

		record.Success = false
		record.ErrMsg = sql.NullString{
			Valid:  true,
			String: errMsg,
		}
	}

	err = db.Save(&record).Error
	if err != nil {
		return err
	}

	return nil
}
