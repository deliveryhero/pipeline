package pipeline

import (
	"context"
)

// Cancel passes an `Item any` from the `in <-chan Item` directly to the out `<-chan Item` until the `Context` is canceled.
// After the context is canceled, everything from `in <-chan Item` is sent to the `cancel` func instead with the `ctx.Err()`.
func Cancel[Item any](ctx context.Context, cancel func(Item, error), in <-chan Item) <-chan Item {
	out := make(chan Item)
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
