package pipeline

import (
	"context"
	"time"
)

// Collect collects `interface{}`s from its in channel and returns `[]interface{}` from its out channel.
// It will collet up to `maxSize` inputs from the `in <-chan interface{}` over up to `maxDuration` before returning them as `[]interface{}`.
// That means when `maxSize` is reached before `maxDuration`, `[maxSize]interface{}` will be passed to the out channel.
// But if `maxDuration` is reached before `maxSize` inputs are collected, `[< maxSize]interface{}` will be passed to the out channel.
// When the `context` is canceled, everything in the buffer will be flushed to the out channel.
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
