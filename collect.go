package pipeline

import (
	"context"
	"time"
)

// Collect collects `interface{}`s from its in channel and returns `[]interface{}` from its out channel.
// It will collect up to `maxSize` inputs from the `in <-chan interface{}` over up to `maxDuration` before returning them as `[]interface{}`.
// That means when `maxSize` is reached before `maxDuration`, `[maxSize]interface{}` will be passed to the out channel.
// But if `maxDuration` is reached before `maxSize` inputs are collected, `[< maxSize]interface{}` will be passed to the out channel.
// When the `context` is canceled, everything in the buffer will be flushed to the out channel.
func Collect(ctx context.Context, maxSize int, maxDuration time.Duration, in <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		for {
			is, open := collect(ctx, maxSize, maxDuration, in)
			if is != nil {
				out <- is
			}
			if !open {
				close(out)
				return
			}
		}
	}()
	return out
}

func collect(ctx context.Context, maxSize int, maxDuration time.Duration, in <-chan interface{}) ([]interface{}, bool) {
	var buffer []interface{}
	timeout := time.After(maxDuration)
	for {
		lenBuffer := len(buffer)
		select {
		case <-ctx.Done():
			// Reduce the timeout to 1/10th of a second
			bs, open := collect(context.Background(), maxSize, 100*time.Millisecond, in)
			return append(buffer, bs...), open
		case <-timeout:
			return buffer, true
		case i, open := <-in:
			if !open {
				return buffer, false
			} else if lenBuffer < maxSize-1 {
				// There is still room in the buffer
				buffer = append(buffer, i)
			} else {
				// There is no room left in the buffer
				return append(buffer, i), true
			}
		}
	}
}
