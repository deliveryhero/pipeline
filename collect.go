package pipeline

import (
	"context"
	"time"
)

// Collect collects `[Item any]`s from its in channel and returns `[]Item` from its out channel.
// It will collect up to `maxSize` inputs from the `in <-chan Item` over up to `maxDuration` before returning them as `[]Item`.
// That means when `maxSize` is reached before `maxDuration`, `[maxSize]Item` will be passed to the out channel.
// But if `maxDuration` is reached before `maxSize` inputs are collected, `[< maxSize]Item` will be passed to the out channel.
// When the `context` is canceled, everything in the buffer will be flushed to the out channel.
func Collect[Item any](ctx context.Context, maxSize int, maxDuration time.Duration, in <-chan Item) <-chan []Item {
	out := make(chan []Item)
	go func() {
		for {
			is, open := collect[Item](ctx, maxSize, maxDuration, in)
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

func collect[Item any](ctx context.Context, maxSize int, maxDuration time.Duration, in <-chan Item) ([]Item, bool) {
	var buffer []Item
	timeout := time.After(maxDuration)
	for {
		lenBuffer := len(buffer)
		select {
		case <-ctx.Done():
			// Reduce the timeout to 1/10th of a second
			bs, open := collect(context.Background(), maxSize-lenBuffer, 100*time.Millisecond, in)
			return append(buffer, bs...), open
		case <-timeout:
			return buffer, true
		case i, open := <-in:
			if !open {
				return buffer, false
			} else if lenBuffer == maxSize-1 {
				// There is no room left in the buffer
				return append(buffer, i), true
			}
			// There is still room in the buffer
			buffer = append(buffer, i)
		}
	}
}
