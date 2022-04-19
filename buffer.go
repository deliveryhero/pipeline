package pipeline

// Buffer creates a buffered channel that will close after the input
// is closed and the buffer is fully drained
func Buffer[Item any](size int, in <-chan Item) <-chan Item {
	buffer := make(chan Item, size)
	go func() {
		for i := range in {
			buffer <- i
		}
		close(buffer)
	}()
	return buffer
}
