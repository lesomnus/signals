package signals_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lesomnus/signals"
)

type env struct {
	s signals.Signal[int]

	ready chan struct{}

	ctx    context.Context
	cancel context.CancelFunc

	wg_p sync.WaitGroup
	wg_c sync.WaitGroup

	num_producers  int
	num_consumers  int
	num_dispatches int
	size_buffer    int
	size_consume   int
	delay_consume  time.Duration
}

func (e *env) Producer() {
	<-e.ready
	for i := range e.num_dispatches {
		if i%100 == 0 {
			for range e.num_consumers {
				e.wg_c.Go(e.Consumer)
			}
		}
		e.s.Dispatch(e.ctx, 42)
	}
}

func (e *env) Consumer() {
	c, close := e.s.Subscribe(e.size_buffer)
	defer close()

	for range e.size_consume {
		select {
		case <-e.ctx.Done():
			return
		case <-time.After(e.delay_consume):
		}

		if _, ok := signals.Recv(e.ctx, c); !ok {
			return
		}
	}
}

func (e *env) Run(b *testing.B) {

	e.ctx, e.cancel = context.WithCancel(context.Background())
	defer e.cancel()

	e.ready = make(chan struct{})
	for range e.num_producers {
		e.wg_p.Go(e.Producer)
	}

	time.Sleep(100 * time.Millisecond)
	close(e.ready)

	b.ResetTimer()
	b.StartTimer()
	e.wg_p.Wait()
	b.StopTimer()

	e.cancel()
	e.wg_c.Wait()
}

func BenchWithSignal(b *testing.B, make_signal func() signals.Signal[int]) {
	delay_base := 100 * time.Microsecond
	b.Run("1->5; no buffer", func(b *testing.B) {
		e := env{
			s:              make_signal(),
			num_producers:  1,
			num_consumers:  5,
			num_dispatches: 4_000,
			size_buffer:    0,
			size_consume:   200,
			delay_consume:  delay_base,
		}
		e.Run(b)
	})
	b.Run("1->5; buffer=100", func(b *testing.B) {
		e := env{
			s:              make_signal(),
			num_producers:  1,
			num_consumers:  5,
			num_dispatches: 4_000,
			size_buffer:    100,
			size_consume:   200,
			delay_consume:  delay_base,
		}
		e.Run(b)
	})
	b.Run("2->5; no buffer", func(b *testing.B) {
		e := env{
			s:              make_signal(),
			num_producers:  2,
			num_consumers:  5,
			num_dispatches: 4_000,
			size_buffer:    0,
			size_consume:   200,
			delay_consume:  delay_base,
		}
		e.Run(b)
	})
	b.Run("2->5; buffer=100", func(b *testing.B) {
		e := env{
			s:              make_signal(),
			num_producers:  2,
			num_consumers:  5,
			num_dispatches: 4_000,
			size_buffer:    100,
			size_consume:   200,
			delay_consume:  delay_base,
		}
		e.Run(b)
	})
	b.Run("4->5; no buffer", func(b *testing.B) {
		e := env{
			s:              make_signal(),
			num_producers:  4,
			num_consumers:  5,
			num_dispatches: 4_000,
			size_buffer:    0,
			size_consume:   200,
			delay_consume:  delay_base,
		}
		e.Run(b)
	})
	b.Run("4->5; buffer=100", func(b *testing.B) {
		e := env{
			s:              make_signal(),
			num_producers:  4,
			num_consumers:  5,
			num_dispatches: 4_000,
			size_buffer:    100,
			size_consume:   200,
			delay_consume:  delay_base,
		}
		e.Run(b)
	})
}

func BenchmarkSoft(b *testing.B) {
	BenchWithSignal(b, signals.Soft[int])
}

func BenchmarkHard(b *testing.B) {
	BenchWithSignal(b, signals.Hard[int])
}

func BenchmarkSure(b *testing.B) {
	BenchWithSignal(b, signals.Sure[int])
}
