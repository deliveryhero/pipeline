package pipeline

// Partition fans a single input channel into multiple output channels based on the partition func.
// Warning: If `partition(interface{}) int` returns an int >= `partitions`, it will cause a panic.
func Partition(partitions int, partition func(interface{}) int, in <-chan interface{}) []<-chan interface{} {
	// Create partitions output channels
	ws := make([]chan<- interface{}, partitions)
	rs := make([]<-chan interface{}, partitions)
	for i := range ws {
		out := make(chan interface{})
		ws[i] = out
		rs[i] = out
	}
	go func() {
		// Route each input into its correct output writer
		for i := range in {
			ws[partition(i)] <- i
		}
		// Close all of the output channels
		for _, w := range ws {
			close(w)
		}
	}()
	// Return the readers
	return rs
}
