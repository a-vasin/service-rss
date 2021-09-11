package signal

import (
	"os"
	"os/signal"
	"syscall"
)

type Handler chan os.Signal

// NewHandler возвращает Handler, который слушает сигналы от ОС.
func NewHandler() Handler {
	handler := Handler(make(chan os.Signal, 1))
	signal.Notify(handler, os.Interrupt, syscall.SIGTERM)
	return handler
}

// Wait ожидает сигнал от ОС
// Блокирует поток выполнения.
func (t Handler) Wait() {
	<-t
	return
}
