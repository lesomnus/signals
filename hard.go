package signals

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
)

// Hard creates a Signal that does not guarantee delivery of all dispatched values to all subscribers.
// Subscribers that cannot keep up with the dispatch rate will be closed by the dispatcher.
func Hard[T any]() Signal[T] {
	s := &signal[T]{
		slots:     atomic.Pointer[[]slot[T]]{},
		make_slot: newHardSlot[T],
	}
	s.slots.Store(&[]slot[T]{})
	return s
}

type hardSlot[T any] struct {
	c chan T

	rw sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	closed bool
}

func newHardSlot[T any](n int) (slot[T], <-chan T) {
	c := make(chan T, n)
	s := &hardSlot[T]{c: c}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s, c
}

func (s *hardSlot[T]) Dispatch(ctx context.Context, v T) (int, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	if s.closed {
		return 0, io.EOF
	}

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-s.ctx.Done():
		return 0, io.EOF
	case s.c <- v:
		return 1, nil
	default:
		s.cancel()
		go s.Close()
		return 0, io.EOF
	}
}

func (s *hardSlot[T]) Close() error {
	s.cancel()

	s.rw.Lock()
	defer s.rw.Unlock()

	if s.closed {
		return nil
	}
	close(s.c)
	s.closed = true

	return nil
}

func (s *hardSlot[T]) Closed() bool {
	return s.ctx.Err() != nil
}
