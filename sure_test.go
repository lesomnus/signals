package signals_test

import (
	"context"
	"testing"
	"time"

	"github.com/lesomnus/signals"
)

func TestSure(t *testing.T) {
	t.Run("no subscribers", func(t *testing.T) {
		s := signals.Sure[int]()

		n, err := s.Dispatch(t.Context(), 42)
		NoError(t, err)
		Equal(t, 0, n)
	})
	t.Run("idle subscriber", func(t *testing.T) {
		t.Parallel()
		s := signals.Sure[int]()

		_, close := s.Subscribe(0)
		defer close()

		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		n, err := s.Dispatch(ctx, 42)
		Equal(t, context.DeadlineExceeded, err)
		Equal(t, 0, n)
	})
	t.Run("queued", func(t *testing.T) {
		s := signals.Sure[int]()

		c, close := s.Subscribe(1)
		defer close()

		n, err := s.Dispatch(t.Context(), 42)
		NoError(t, err)
		Equal(t, 1, n)

		v := <-c
		Equal(t, 42, v)
	})
	t.Run("multiple subscribers", func(t *testing.T) {
		t.Parallel()
		s := signals.Sure[int]()

		c1, close1 := s.Subscribe(1)
		defer close1()
		_, close2 := s.Subscribe(0)
		defer close2()
		c3, close3 := s.Subscribe(1)
		defer close3()

		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		n, err := s.Dispatch(ctx, 42)
		Equal(t, context.DeadlineExceeded, err)
		Equal(t, 1, n)

		v1 := <-c1
		Equal(t, 42, v1)

		select {
		case <-c3:
			t.Fatal("expected no value to be dispatched to c3")
		default:
		}
	})
}
