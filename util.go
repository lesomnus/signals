package signals

import "context"

func Recv[T any](ctx context.Context, c <-chan T) (v T, ok bool) {
	select {
	case <-ctx.Done():
		return v, false
	case v, ok = <-c:
		return v, ok
	}
}
