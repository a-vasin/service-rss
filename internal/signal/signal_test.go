package signal

import (
	"syscall"
	"testing"
	"time"
)

func TestHandler(t *testing.T) {
	timeout := time.After(1 * time.Second)
	done := make(chan bool)

	go func() {
		h := NewHandler()
		h <- syscall.SIGTERM
		h.Wait()

		done <- true
	}()

	select {
	case <-timeout:
		t.Fatal("test didn't finish in time")
	case <-done:
	}
}
