// Copyright 2025 AnonymousMessage Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package httpserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"github.com/SongZihuan/anonymous-message/src/httpserver/engine"
	"github.com/pires/go-proxyproto"
	"net"
	"net/http"
	"time"
)

var server *http.Server

func InitHttpSystem() (chan bool, error) {
	if server != nil {
		return nil, fmt.Errorf("http server is running")
	}

	err := engine.InitEngine()
	if err != nil {
		return nil, err
	}

	server = &http.Server{
		Addr:    flagparser.HttpAddress,
		Handler: engine.Engine,
	}

	tcpListener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %s\n", server.Addr, err.Error())
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

	var httpchan = make(chan bool)

	go func() {
		defer close(httpchan)

		fmt.Printf("Http Server start on: %s\n", server.Addr)
		err := server.Serve(listener)
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				fmt.Printf("Http Server stop on: %s\n", server.Addr)
			} else {
				fmt.Printf("Http Server stop on: %s by error: %s\n", server.Addr, err.Error())
			}
		}

		server = nil
	}()

	return httpchan, nil
}

func ShutdownHttpSystem() {
	if server == nil {
		return
	}

	defer func() {
		server = nil
	}()

	ctx, fn := context.WithTimeout(context.Background(), 10*time.Second)
	defer fn()

	_ = server.Shutdown(ctx)
}
