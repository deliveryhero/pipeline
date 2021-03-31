package util

import (
	"context"
	"time"
)

// Collect collects up to maxSize inputs over up to maxDuration before returning them as []interface{}.
// If maxSize is reached before maxDuration, [maxSize]interface{} will be returned.
// If maxDuration is reached before maxSize is collected, [>maxSize]interface{} will be returned.
// If no inputs are collected over maxDuration, no outputs will be returned.
// If the context is closed, everything in the buffer will be immediately flushed and there will be no delay for any remaining messages
func Collect(ctx context.Context, maxSize int, maxDuration time.Duration, in <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		var buffer []interface{}
		timeout := time.After(maxDuration)
		for {
			lenBuffer := len(buffer)
			select {
			case <-ctx.Done():
				if lenBuffer > 0 {
					out <- buffer
					buffer = nil
				}
				timeout = nil
			case i, open := <-in:
				if !open && lenBuffer > 0 {
					// We have some interfaces left to to return when in is closed
					out <- buffer
					return
				} else if !open {
					return
				} else if lenBuffer < maxSize-1 {
					// There is still room in the buffer
					buffer = append(buffer, i)
				} else {
					// There is no room left in the buffer
					out <- append(buffer, i)
					buffer = nil
					timeout = time.After(maxDuration)
				}
			case <-timeout:
				if lenBuffer > 0 {
					// We timed out with some items left in the buffer
					out <- buffer
					buffer = nil
				}
				timeout = time.After(maxDuration)
			}
		}
	}()
	return out
}
