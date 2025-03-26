// Copyright 2025 AnonymousMessage Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package emailserver

import (
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/emailserver/emailaddress"
	"github.com/SongZihuan/anonymous-message/src/emailserver/imapserver"
	"github.com/SongZihuan/anonymous-message/src/emailserver/smtpserver"
)

var ready = false
var imapStopChannel chan bool

func InitEmailSystem() (err error) {
	defer func() {
		if err == nil {
			ready = true
		}
	}()

	err = smtpserver.InitSmtp()
	if err != nil {
		return err
	}

	err = imapserver.InitImap()
	if err != nil {
		return err
	}

	err = emailaddress.InitGlobalAddress()
	if err != nil {
		return err
	}

	return nil
}

func StartEmailServer() (chan bool, error) {
	if !ready {
		return nil, fmt.Errorf("email server is not ready")
	} else if imapStopChannel != nil {
		return nil, fmt.Errorf("email server is running")
	}

	imapStopChannel = make(chan bool)

	imapchan, err := imapserver.StartIMAPServer(imapStopChannel)
	if err != nil {
		close(imapStopChannel)
		imapStopChannel = nil
		return nil, err
	}

	return imapchan, nil
}

func StopEmailServer() error {
	if !ready {
		return fmt.Errorf("email server is not ready")
	} else if imapStopChannel == nil {
		return fmt.Errorf("email server is not runned")
	}

	close(imapStopChannel)
	imapStopChannel = nil

	return nil
}
