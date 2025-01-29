package server

import (
	"errors"
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/engine"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"github.com/SongZihuan/anonymous-message/src/iprate"
	"github.com/SongZihuan/anonymous-message/src/signalchan"
	"net/http"
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

	err = iprate.InitRedis()
	if err != nil {
		fmt.Printf("init redis fail: %s\n", err.Error())
		return 1
	}
	defer iprate.CloseRedis()

	err = engine.InitEngine()
	if err != nil {
		fmt.Printf("init engine fail: %s\n", err.Error())
		return 1
	}

	err = signalchan.InitSignal()
	if err != nil {
		fmt.Printf("init signal fail: %s\n", err.Error())
		return 1
	}
	defer signalchan.CloseSignal()

	var httpchan = make(chan error)
	defer func() {
		close(httpchan)
	}()

	go func() {
		fmt.Printf("Http Server start on: %s\n", flagparser.HttpAddress)
		httpchan <- http.ListenAndServe(flagparser.HttpAddress, engine.Engine)
	}()

	select {
	case <-signalchan.SignalChan:
		fmt.Printf("Server closed: safe\n")
		return 0
	case err := <-httpchan:
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Http Server closed: safe\n")
			return 0
		}
		fmt.Printf("Http Server error closed: %s\n", err.Error())
		return 1
	}
	// 后续不可达
}
