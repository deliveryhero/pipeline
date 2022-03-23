package pipeline

// Drain empties the input and blocks until the channel is closed
func Drain[Item any](in <-chan Item) {
	for range in {
	}
}
