# signals

Please bench signals in your application to see if they are a good fit for your use case.
Note that `Soft` is extremely slowed down when there are many producers.

## Usage

```go
package main

import (
	"fmt"

	"github.com/lesomnus/signals"
)

func main() {
	s := signals.Sure[int]()
	c, close := s.Subscribe(10)
	defer close()
	
	go func() {
		for i := 0; i < 100; i++ {
			s.Dispatch(i)
		}
	}()
	
	for v := range c {
		fmt.Println(v)
	}
}
```

- `Soft` does NOT guarantee delivery of all dispatched values; it *skips* the slot if the subscriber is not ready to receive.
- `Hard` does NOT guarantee delivery of all dispatched values; it *closes* the slot if the subscriber is not ready to receive.
- `Sure` guarantees delivery of all dispatched values; it blocks the dispatcher until the subscriber is ready to receive.
