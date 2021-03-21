package util

import "context"

// Cacel passes the in chan to the out chan until the context is canceled.
// When the context is canceled, everything from in is passed to the cancel func instead of out.
// Out closes when in is closed.
func Cancel(ctx context.Context, canceled func(i interface{}), in <-chan interface{}) <-chan interface{} {
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
					canceled(i)
				}
				return
			}
		}
	}()
	return out
}
