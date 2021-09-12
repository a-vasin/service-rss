package signal

import (
	"os"
	"os/signal"
	"syscall"
)

type Handler chan os.Signal

func NewHandler() Handler {
	handler := Handler(make(chan os.Signal, 1))
	signal.Notify(handler, os.Interrupt, syscall.SIGTERM)
	return handler
}

func (t Handler) Wait() {
	<-t
	return
}
