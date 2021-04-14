package pipeline

import "context"

// Cacel passes `interface{}`s from the in chan to the out chan until the context is canceled.
// When the context is canceled, everything from in is intercepted by the `cancel` func instead of being passed to the out chan.
func Cancel(ctx context.Context, cancel func(interface{}, error), in <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		for {
			select {
			// When the context isnt canceld, pass everything to the out chan
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
