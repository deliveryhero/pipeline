package util

// Concurrently fans a single channel out to multiple channels, all listening to the in channel concurrently
func Concurrently(count int, in <-chan interface{}) (outs []<-chan interface{}) {

	// Listen to in count times, concurrently
	for i := 0; i < count; i++ {

		// Create a seprate out chan for each concurrent execution
		out := make(chan interface{})
		outs = append(outs, out)

		go func() {
			// Wait for in to close
			for i := range in {
				if i != nil {
					// Fan the in chan into one of many out chans
					out <- i
				}
			}
			// Close all outs cooresponding to the in
			close(out)
		}()
	}
	return
}
