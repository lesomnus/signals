package signals

import (
	"context"
	"sync/atomic"
)

// Soft creates a Signal that does not guarantee delivery of all dispatched values to all subscribers.
// Subscribers that cannot keep up with the dispatch rate will be skipped until they can receive the value.
func Soft[T any]() Signal[T] {
	s := &signal[T]{
		slots:     atomic.Pointer[[]slot[T]]{},
		make_slot: newSoftSlot[T],
	}
	s.slots.Store(&[]slot[T]{})
	return s
}

type softSlot[T any] struct {
	c chan T

	is_closed atomic.Bool
}

func newSoftSlot[T any](n int) (slot[T], <-chan T) {
	c := make(chan T, n)
	s := &softSlot[T]{c: c}

	return s, c
}

func (s *softSlot[T]) Dispatch(ctx context.Context, v T) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case s.c <- v:
		return 1, nil
	default:
		return 0, nil
	}
}

func (s *softSlot[T]) Close() error {
	s.is_closed.Store(true)
	return nil
}

func (s *softSlot[T]) Closed() bool {
	return s.is_closed.Load()
}
