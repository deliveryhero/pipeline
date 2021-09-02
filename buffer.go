package pipeline

// Buffer creates a buffered channel that will close after the input
// is closed and the buffer is fully drained
func Buffer(size int, in <-chan interface{}) <-chan interface{} {
	buffer := make(chan interface{}, size)
	go func() {
		for i := range in {
			buffer <- i
		}
		close(buffer)
	}()
	return buffer
}
