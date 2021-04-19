package pipeline

import (
	"context"
)

// Cancel passes an `interface{}` from the `in <-chan interface{}` directly to the out `<-chan interface{}` until the `Context` is canceled.
// After the context is canceled, everything from `in <-chan interface{}` is sent to the `cancel` func instead with the `ctx.Err()`.
func Cancel(ctx context.Context, cancel func(interface{}, error), in <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		for {
			select {
			// When the context isn't canceld, pass everything to the out chan
			// until in is closed
			case i, open := <-in:
				if !open {
					return
				}
				out <- i
			// When the context is canceled, pass all ins to the
			// cancel fun until in is closed
			case <-ctx.Done():
				for i := range in {
					cancel(i, ctx.Err())
				}
				return
			}
		}
	}()
	return out
}
