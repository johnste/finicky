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

func TestQueueURLNeverBlocksOnAFullQueue(t *testing.T) {
	done := make(chan struct{})
	go func() {
		for i := 0; i <= cap(urlListener); i++ {
			queueURL(URLInfo{URL: "https://example.com"})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("queueing a url blocked on a full queue")
	}

	if len(urlListener) != cap(urlListener) {
		t.Errorf("queued %d urls, want the queue full at %d", len(urlListener), cap(urlListener))
	}
	for len(urlListener) > 0 {
		<-urlListener
	}
}
