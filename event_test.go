package signals_test

import (
	"testing"

	"github.com/lesomnus/signals"
)

func TestEvent(t *testing.T) {
	storage := map[int][]int{}
	put := func(k, v int) {
		vs := storage[k]
		vs = append(vs, v)
		storage[k] = vs
	}

	event := signals.NewEvent[int]()
	close0 := event.Listen(func(v int) { put(0, v) })
	defer close0()
	close1 := event.Listen(func(v int) { put(1, v) })
	defer close1()

	event.Emit(42)
	EqualSlice(t, storage[0], []int{42})
	EqualSlice(t, storage[1], []int{42})

	close0()
	event.Emit(43)
	EqualSlice(t, storage[0], []int{42})
	EqualSlice(t, storage[1], []int{42, 43})

	close2 := event.Listen(func(v int) { put(2, v) })
	defer close2()

	event.Emit(44)
	EqualSlice(t, storage[0], []int{42})
	EqualSlice(t, storage[1], []int{42, 43, 44})
	EqualSlice(t, storage[2], []int{44})

	close1()
	event.Emit(45)
	EqualSlice(t, storage[0], []int{42})
	EqualSlice(t, storage[1], []int{42, 43, 44})
	EqualSlice(t, storage[2], []int{44, 45})

	close2()
	event.Emit(46)
	EqualSlice(t, storage[0], []int{42})
	EqualSlice(t, storage[1], []int{42, 43, 44})
	EqualSlice(t, storage[2], []int{44, 45})
}
