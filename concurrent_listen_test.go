package signals_test

import (
	"sync"
	"testing"

	"github.com/lesomnus/signals"
)

func TestListenConcurrently(t *testing.T) {
	e := signals.NewEvent[int]()

	// One first, so that the stored slice has spare capacity for the rest to
	// append into -- which is the shape the bug needed.
	e.Listen(func(int) {})

	const n = 8
	var (
		wg    sync.WaitGroup
		stops = make([]func(), n)
	)
	for i := range stops {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stops[i] = e.Listen(func(int) {})
		}()
	}
	wg.Wait()

	if got := e.Emit(1); got != n+1 {
		t.Fatalf("want %d listeners, got %d", n+1, got)
	}

	// And every one of them can stop, which is what fails when a listener was
	// published into a list it is not in.
	for _, f := range stops {
		f()
	}
	if got := e.Emit(1); got != 1 {
		t.Fatalf("want 1 listener left, got %d", got)
	}
}

func TestSubscribeConcurrently(t *testing.T) {
	for _, tc := range []struct {
		desc string
		s    signals.Signal[int]
	}{
		{"soft", signals.Soft[int]()},
		{"hard", signals.Hard[int]()},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			const n = 8

			var wg sync.WaitGroup
			for range n {
				wg.Add(1)
				go func() {
					defer wg.Done()
					tc.s.Subscribe(1)
				}()
			}
			wg.Wait()

			got, err := tc.s.Dispatch(t.Context(), 1)
			if err != nil {
				t.Fatal(err)
			}
			if got != n {
				t.Fatalf("want %d subscribers, got %d", n, got)
			}
		})
	}
}
