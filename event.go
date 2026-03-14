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
		next := append(*curr, fp)
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
			next := make([]*Callback[T], len(*curr)-1)

			i := 0
			for _, fp_ := range *curr {
				if fp_ == fp {
					continue
				}

				next[i] = fp_
				i++
			}
			if e.callbacks.CompareAndSwap(curr, &next) {
				break
			}
		}
	}
}
