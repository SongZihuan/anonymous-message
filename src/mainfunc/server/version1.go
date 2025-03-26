package server

import (
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/database"
	"github.com/SongZihuan/anonymous-message/src/emailserver"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"github.com/SongZihuan/anonymous-message/src/httpserver"
	"github.com/SongZihuan/anonymous-message/src/reqrate"
	"github.com/SongZihuan/anonymous-message/src/signalchan"
	"time"
)

func MainV1() (exitcode int) {
	defer func() {
		if recover() != nil {
			exitcode = 1
			return
		}
	}()

	var hasPrint = false
	err := flagparser.InitFlagParser()
	if err != nil {
		fmt.Printf("init flag fail: %s\n", err.Error())
		return 1
	}

	if flagparser.Version {
		_, _ = flagparser.PrintVersion()
		hasPrint = true
		return 0
	}

	if flagparser.License {
		if hasPrint {
			_, _ = flagparser.PrintLF()
		}
		_, _ = flagparser.PrintLicense()
		hasPrint = true
		return 0
	}

	if flagparser.Report {
		if hasPrint {
			_, _ = flagparser.PrintLF()
		}
		_, _ = flagparser.PrintReport()
		hasPrint = true
		return 0
	}

	if flagparser.DryRun || flagparser.ShowOption {
		if hasPrint {
			_, _ = flagparser.PrintLF()
		}

		flagparser.Print()
	}

	if flagparser.DryRun {
		return 0
	}

	defer func() {
		time.Sleep(1 * time.Second)
	}()

	err = reqrate.RateClean()
	if err != nil {
		fmt.Printf("init rate clean fail: %s\n", err.Error())
		return 1
	}

	err = database.InitSQLite()
	if err != nil {
		fmt.Printf("init sqlite fail: %s\n", err.Error())
		return 1
	}
	defer database.CloseSQLite()

	err = emailserver.InitEmailSystem()
	if err != nil {
		fmt.Printf("init email system fail: %s\n", err.Error())
		return 1
	}

	imapchan, err := emailserver.StartEmailServer()
	if err != nil {
		fmt.Printf("init email server (imap) fail: %s\n", err.Error())
		return 1
	}
	defer func() {
		_ = emailserver.StopEmailServer()
	}()

	httpchan, err := httpserver.InitHttpSystem()
	if err != nil {
		fmt.Printf("init http server fail: %s\n", err.Error())
		return 1
	}
	defer func() {
		httpserver.ShutdownHttpSystem()
	}()

	err = signalchan.InitSignal()
	if err != nil {
		fmt.Printf("init signal fail: %s\n", err.Error())
		return 1
	}
	defer signalchan.CloseSignal()

	select {
	case <-signalchan.SignalChan:
		fmt.Printf("Server safe closed by signal.\n")
		return 0
	case <-httpchan:
		fmt.Printf("Server safe closed http server.\n")
		return 0
	case <-imapchan:
		fmt.Printf("Server safe closed email server.\n")
		return 0
	}
	// 后续不可达
}
