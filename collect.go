package util

import (
	"time"
)

// Collect outputs `max` interfaces at a time from `in` or whatever came through during the last `duration`
func Collect(max int, duration time.Duration, in <-chan interface{}) <-chan []interface{} {
	out := make(chan []interface{})
	go func() {
		// Correct memory management
		defer close(out)

		var buffer []interface{}
		timeout := time.After(duration)
		for {
			lenBuffer := len(buffer)
			select {
			case i, open := <-in:
				if !open && lenBuffer > 0 {
					// We have some interfaces left to to return when in is closed
					out <- buffer
					return
				} else if !open {
					return
				} else if lenBuffer < max-1 {
					// There is still room in the buffer
					buffer = append(buffer, i)
				} else {
					// There is no room left in the buffer
					out <- append(buffer, i)
					buffer = nil
				}
			case <-timeout:
				if lenBuffer > 0 {
					// We timed out with some items left in the buffer
					out <- buffer
					buffer = nil
					timeout = time.After(duration)
				}
			}
		}
	}()
	return out
}
