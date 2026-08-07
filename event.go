package signals

import (
	"sync/atomic"
)

type Event[T any] interface {
	Emitter[T]
	Listener[T]
}

type Emitter[T any] interface {
	Emit(v T) int
}

type Listener[T any] interface {
	Listen(f Callback[T]) func()
}

type Callback[T any] func(v T)

// NewEvent creates an Event for multiple emitter.
func NewEvent[T any]() Event[T] {
	e := &event[T]{}

	fs := []*Callback[T]{}
	e.callbacks.Store(&fs)

	return e
}

type event[T any] struct {
	callbacks atomic.Pointer[[]*Callback[T]]
}

func (e *event[T]) Emit(v T) int {
	fs := *e.callbacks.Load()
	for _, f := range fs {
		(*f)(v)
	}
	return len(fs)
}

func (e *event[T]) Listen(f Callback[T]) func() {
	fp := &f
	for {
		curr := e.callbacks.Load()

		// A copy, and never append(*curr, fp). The stored slice usually has
		// spare capacity, so appending to it writes into the array Emit is
		// ranging over -- and two Listen calls that race write the same index,
		// so the loser of the compare-and-swap has already overwritten the
		// winner's callback. What gets published is then a list one listener is
		// missing from, and that listener's unsubscribe has nothing to remove.
		next := make([]*Callback[T], len(*curr), len(*curr)+1)
		copy(next, *curr)
		next = append(next, fp)

		if e.callbacks.CompareAndSwap(curr, &next) {
			break
		}
	}

	var closed atomic.Bool
	return func() {
		if closed.Swap(true) {
			return
		}

		for {
			curr := e.callbacks.Load()

			// Appended into an empty slice rather than written at an index of a
			// slice one shorter. This is right whether the callback is in the
			// list or not, and the other way round panics on a list that does
			// not hold it -- which is reachable, since a caller may hold on to
			// the function through anything that rebuilds the list.
			next := make([]*Callback[T], 0, len(*curr))
			for _, fp_ := range *curr {
				if fp_ == fp {
					continue
				}

				next = append(next, fp_)
			}
			if e.callbacks.CompareAndSwap(curr, &next) {
				break
			}
		}
	}
}
