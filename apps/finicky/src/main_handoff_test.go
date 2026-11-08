package main

import (
	"testing"
	"time"
)

// The main thread calls these while the url loop can be busy in the picker,
// waiting on the main thread itself, so they must return without a receiver.
func TestMainThreadHandoffsNeverBlock(t *testing.T) {
	done := make(chan struct{})
	go func() {
		QueueWindowDisplay(1)
		QueueWindowDisplay(0)
		WindowDidClose()
		WindowDidClose()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("a main thread hand-off blocked with nobody receiving")
	}

	if got := <-queueWindowOpen; !got {
		t.Errorf("queued window display = %v, want the first value, true", got)
	}
	<-windowClosed
}
