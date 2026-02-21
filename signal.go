package signals

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
)

type Signal[T any] interface {
	Dispatcher[T]
	Subscriber[T]
}

type Dispatcher[T any] interface {
	Dispatch(ctx context.Context, v T) (int, error)
}

type Subscriber[T any] interface {
	// Subscribe returns a channel that receives dispatched values and a function to close the subscription.
	// For Sure and Sort, closing a subscription does not close the channel so channel must be selected with a context cancellation or other channel to avoid deadlock.
	// For Hard, closing a subscription will close the channel and make the subscriber stop receiving values.
	// Use [Recv] to receive values from the channel with context for convenience.
	Subscribe(n int) (<-chan T, func() error)
}

type signal[T any] struct {
	slots     atomic.Pointer[[]slot[T]]
	make_slot func(n int) (slot[T], <-chan T)
}

func (s *signal[T]) Dispatch(ctx context.Context, v T) (int, error) {
	slots := s.slots.Load()
	if slots == nil {
		return 0, nil
	}

	var count int
	for _, slot := range *slots {
		n, err := slot.Dispatch(ctx, v)
		count += n

		if err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}
			return count, err
		}
	}

	return count, nil
}

func (s *signal[T]) Subscribe(n int) (<-chan T, func() error) {
	v, c := s.make_slot(n)

	next := []slot[T]{}
	for {
		curr := s.slots.Load()
		if cap(next) < len(*curr)+1 {
			next = make([]slot[T], 0, len(*curr)+1)
		} else {
			next = next[:0]
		}

		for _, slot := range *curr {
			if !slot.Closed() {
				next = append(next, slot)
			}
		}

		next = append(next, v)
		if s.slots.CompareAndSwap(curr, &next) {
			break
		}
	}

	return c, v.Close
}

type slot[T any] interface {
	Dispatcher[T]
	io.Closer
	Closed() bool
}
