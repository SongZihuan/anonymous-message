package signalchan

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var SignalChan = make(chan os.Signal)

func InitSignal() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("init signal error: %v", r)
			return
		}
	}()

	signal.Notify(SignalChan, syscall.SIGINT, syscall.SIGTERM)
	return nil
}

func CloseSignal() {
	signal.Stop(SignalChan)
	close(SignalChan)
}
