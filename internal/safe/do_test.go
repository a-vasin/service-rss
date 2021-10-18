package safe

import (
	"sync"
	"testing"
)

func TestDo(t *testing.T) {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go Do(func() {
		defer wg.Done()
		panic("test panic")
	})

	wg.Wait()
}
