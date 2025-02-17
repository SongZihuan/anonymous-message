package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/database"
	"github.com/SongZihuan/anonymous-message/src/email/imapserver"
	"github.com/SongZihuan/anonymous-message/src/email/smtpserver"
	"github.com/SongZihuan/anonymous-message/src/engine"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"github.com/SongZihuan/anonymous-message/src/reqrate"
	"github.com/SongZihuan/anonymous-message/src/signalchan"
	"github.com/SongZihuan/anonymous-message/src/utils"
	"github.com/pires/go-proxyproto"
	"net"
	"net/http"
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

	err = reqrate.InitRedis()
	if err != nil {
		fmt.Printf("init redis fail: %s\n", err.Error())
		return 1
	}
	defer reqrate.CloseRedis()

	err = database.InitSQLite()
	if err != nil {
		fmt.Printf("init sqlite fail: %s\n", err.Error())
		return 1
	}
	defer database.CloseSQLite()

	err = smtpserver.InitSmtp()
	if err != nil {
		fmt.Printf("init smtp server fail: %s\n", err.Error())
		return 1
	}

	imapstopchan, err := imapserver.StartIMAPServer()
	if err != nil {
		fmt.Printf("init imap fail: %s\n", err.Error())
		return 1
	}
	defer func() {
		if imapstopchan != nil {
			close(imapstopchan)
			imapstopchan = nil
		}
	}()

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

	server := http.Server{
		Addr:    flagparser.HttpAddress,
		Handler: engine.Engine,
	}

	tcpListener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		fmt.Printf("listen on %s: %s\n", server.Addr, err.Error())
		return 1
	}

	var listener net.Listener
	if !flagparser.NotProxyProto {
		proxyListener := &proxyproto.Listener{
			Listener:          tcpListener,
			ReadHeaderTimeout: 10 * time.Second,
		}
		listener = proxyListener
	} else {
		listener = tcpListener
	}
	defer func() {
		_ = listener.Close()
	}()

	var httpchan = make(chan error)
	defer func() {
		close(httpchan)
		httpchan = nil
	}()

	go func() {
		fmt.Printf("Http Server start on: %s\n", server.Addr)
		err := server.Serve(listener)
		if utils.IsChanOpen(httpchan) {
			httpchan <- err
		}
	}()

	defer func() {
		time.Sleep(1 * time.Second)
	}()

	select {
	case <-signalchan.SignalChan:
		fmt.Printf("Server closed: safe\n")
		if utils.IsChanOpen(imapstopchan) {
			imapstopchan <- true
		}

		func() {
			ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelFunc()

			_ = server.Shutdown(ctx)
		}()
		return 0
	case err := <-httpchan:
		if utils.IsChanOpen(imapstopchan) {
			imapstopchan <- true
		}

		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Http Server closed: safe\n")
			return 0
		}

		fmt.Printf("Http Server error closed: %s\n", err.Error())
		return 1
	}
	// 后续不可达
}
