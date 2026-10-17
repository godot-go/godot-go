package core

import (
	"testing"
	"time"
)

func TestSignalAwaitStateDeliverSendsPayloadThenCloses(t *testing.T) {
	// Catches: deliver() forgetting the send (receive blocks) or forgetting
	// the close (second receive / range would hang).
	s := newAwaitState()
	args := []any{int64(4), "critical"}

	s.deliver(args)

	got, ok := <-s.results
	if !ok {
		t.Fatal("expected delivered result, channel closed without a send")
	}
	if len(got.Args) != 2 || got.Args[0] != int64(4) || got.Args[1] != "critical" {
		t.Fatalf("expected args [4 critical], got %v", got.Args)
	}
	if _, alive := <-s.results; alive {
		t.Fatal("expected channel closed after delivery")
	}
}

func TestSignalAwaitStateDissolveClosesWithoutSend(t *testing.T) {
	// Catches: dissolve() sending a spurious value (ok would be true) or not
	// closing at all (receive blocks).
	s := newAwaitState()

	s.dissolve()

	got, ok := <-s.results
	if ok {
		t.Fatalf("expected zero result with ok=false, got %+v", got)
	}
	if got.Args != nil {
		t.Fatalf("expected zero-value result, got %+v", got)
	}
}

func TestSignalAwaitStateSecondTerminalPathIsNoOp(t *testing.T) {
	// Catches: a missing sync.Once — whichever second path runs (deliver or
	// dissolve) would panic with "close of closed channel".
	t.Run("dissolveAfterDeliverKeepsPayload", func(t *testing.T) {
		s := newAwaitState()
		s.deliver([]any{int64(1)})
		s.dissolve()
		got, ok := <-s.results
		if !ok || got.Args[0] != int64(1) {
			t.Fatalf("expected first delivery to win, got %+v ok=%v", got, ok)
		}
	})
	t.Run("deliverAfterDissolveCannotResurrect", func(t *testing.T) {
		s := newAwaitState()
		s.dissolve()
		s.deliver([]any{int64(99)})
		got, ok := <-s.results
		if ok {
			t.Fatalf("expected closed-without-send to be final, got %+v", got)
		}
	})
	t.Run("deliverTwiceKeepsFirst", func(t *testing.T) {
		s := newAwaitState()
		s.deliver([]any{int64(1)})
		s.deliver([]any{int64(2)})
		got, ok := <-s.results
		if !ok || got.Args[0] != int64(1) {
			t.Fatalf("expected first delivery to win, got %+v ok=%v", got, ok)
		}
	})
}

func TestSignalAwaitStateDeliverNeverBlocksWithoutReceiver(t *testing.T) {
	// Catches: an unbuffered channel — deliver would block until a receiver
	// arrives, deadlocking the emitting engine thread in production.
	s := newAwaitState()
	done := make(chan struct{})
	go func() {
		s.deliver([]any{int64(7)})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("deliver blocked with no receiver waiting")
	}
}
