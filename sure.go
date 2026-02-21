package signals

import (
	"context"
	"io"
	"sync/atomic"
)

// Sure creates a Signal that guarantees delivery of all dispatched values to all subscribers.
// Subscribers that cannot keep up with the dispatch rate will block the dispatching until they can receive the value.
func Sure[T any]() Signal[T] {
	s := &signal[T]{
		slots:     atomic.Pointer[[]slot[T]]{},
		make_slot: newSureSlot[T],
	}
	s.slots.Store(&[]slot[T]{})
	return s
}

type sureSlot[T any] struct {
	c chan T

	ctx    context.Context
	cancel context.CancelFunc
}

func newSureSlot[T any](n int) (slot[T], <-chan T) {
	c := make(chan T, n)
	s := &sureSlot[T]{c: c}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s, c
}

func (s *sureSlot[T]) Dispatch(ctx context.Context, v T) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-s.ctx.Done():
		return 0, io.EOF
	case s.c <- v:
		return 1, nil
	}
}

func (s *sureSlot[T]) Close() error {
	s.cancel()
	return nil
}

func (s *sureSlot[T]) Closed() bool {
	return s.ctx.Err() != nil
}
