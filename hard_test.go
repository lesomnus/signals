package signals_test

import (
	"testing"

	"github.com/lesomnus/signals"
)

func TestHard(t *testing.T) {
	t.Run("no subscribers", func(t *testing.T) {
		s := signals.Hard[int]()

		n, err := s.Dispatch(t.Context(), 42)
		NoError(t, err)
		Equal(t, 0, n)
	})
	t.Run("idle subscriber", func(t *testing.T) {
		s := signals.Hard[int]()

		_, close := s.Subscribe(0)
		defer close()

		n, err := s.Dispatch(t.Context(), 42)
		NoError(t, err)
		Equal(t, 0, n)
	})
	t.Run("queued", func(t *testing.T) {
		s := signals.Hard[int]()

		c, close := s.Subscribe(1)
		defer close()

		n, err := s.Dispatch(t.Context(), 42)
		NoError(t, err)
		Equal(t, 1, n)

		v := <-c
		Equal(t, 42, v)
	})
	t.Run("multiple subscribers", func(t *testing.T) {
		s := signals.Hard[int]()

		c1, close1 := s.Subscribe(1)
		defer close1()
		c2, close2 := s.Subscribe(0)
		defer close2()
		c3, close3 := s.Subscribe(1)
		defer close3()

		n, err := s.Dispatch(t.Context(), 42)
		NoError(t, err)
		Equal(t, 2, n)

		v1 := <-c1
		Equal(t, 42, v1)
		_, ok := <-c2
		Equal(t, false, ok)
		v3 := <-c3
		Equal(t, 42, v3)
	})
}
