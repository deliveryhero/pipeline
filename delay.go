package pipeline

import (
	"context"
	"time"
)

// Delay delays reading each input by `duration`.
// If the context is canceled, the delay will not be applied.
func Delay(ctx context.Context, duration time.Duration, in <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		// Keep reading from in until its closed
		for i := range in {
			// Take one element from in and pass it to out
			out <- i
			select {
			// Wait duration before reading another input
			case <-time.After(duration):
			// Don't wait if the context is canceled
			case <-ctx.Done():
			}
		}
	}()
	return out
}
