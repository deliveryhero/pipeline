package pipeline

import "sync"

// Merge fans multiple channels in to a single channel
func Merge(ins ...<-chan interface{}) <-chan interface{} {
	// Don't merge anything if we don't have to
	if l := len(ins); l == 0 {
		out := make(chan interface{})
		close(out)
		return out
	} else if l == 1 {
		return ins[0]
	}
	out := make(chan interface{})
	// Create a WaitGroup that waits for all of the ins to close
	var wg sync.WaitGroup
	wg.Add(len(ins))
	go func() {
		// When all of the ins are closed, close the out
		wg.Wait()
		close(out)
	}()
	for i := range ins {
		go func(in <-chan interface{}) {
			// Wait for each in to close
			for i := range in {
				if i != nil {
					// Fan the contents of each in into the out
					out <- i
				}
			}
			// Tell the WaitGroup that one of the channels is closed
			wg.Done()
		}(ins[i])
	}
	return out
}
