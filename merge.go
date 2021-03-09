package util

import (
	"sync"
)

// Merge fans multiple error channels in to a single error channel
func Merge(errChans ...<-chan error) <-chan error {
	mergedChan := make(chan error)

	// Create a WaitGroup that waits for all of the errChans to close
	var wg sync.WaitGroup
	wg.Add(len(errChans))
	go func() {
		// When all of the errChans are closed, close the mergedChan
		wg.Wait()
		close(mergedChan)
	}()

	for i := range errChans {
		go func(errChan <-chan error) {
			// Wait for each errChan to close
			for err := range errChan {
				if err != nil {
					// Fan the contents of each errChan into the mergedChan
					mergedChan <- err
				}
			}
			// Tell the WaitGroup that one of the errChans is closed
			wg.Done()
		}(errChans[i])
	}

	return mergedChan
}
