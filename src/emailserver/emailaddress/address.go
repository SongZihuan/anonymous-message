// Copyright 2025 AnonymousMessage Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package emailaddress

import (
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"net/mail"
)

var DefaultRecipientAddress *mail.Address

var NoticeAddressList []*mail.Address
var NoticeAddress map[string]*mail.Address

var RecipientAddress map[string]*mail.Address

func InitGlobalAddress() (err error) {
	DefaultRecipientAddress = &mail.Address{
		Name:    flagparser.Name,
		Address: flagparser.SMTPUser,
	}

	NoticeAddressList, err = mail.ParseAddressList(flagparser.NoticeList)
	if err != nil {
		return fmt.Errorf("parser notice email address list fail: %s", err.Error())
	}

	if len(NoticeAddressList) == 0 {
		return fmt.Errorf("notice address list is empty")
	}

	NoticeAddress = make(map[string]*mail.Address, len(NoticeAddressList))

	for _, address := range NoticeAddressList {
		NoticeAddress[address.Address] = address
	}

	if len(flagparser.RecipientList) == 0 {
		RecipientAddress = make(map[string]*mail.Address, 1)

		RecipientAddress[flagparser.IMAPUser] = &mail.Address{
			Name:    flagparser.Name,
			Address: flagparser.IMAPUser,
		}
	} else {
		recipientList, err := mail.ParseAddressList(flagparser.RecipientList)
		if err != nil {
			return fmt.Errorf("parser recipient email address list fail: %s", err.Error())
		}

		RecipientAddress = make(map[string]*mail.Address, len(recipientList)+1)

		RecipientAddress[flagparser.IMAPUser] = &mail.Address{
			Name:    flagparser.Name,
			Address: flagparser.IMAPUser,
		}

		for _, address := range recipientList {
			RecipientAddress[address.Address] = address
		}
	}

	return nil
}
