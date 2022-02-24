package pipeline

// Emit fans `is ...Item`` out to a `<-chan Item`
func Emit[Item any](is ...Item) <-chan Item {
	out := make(chan Item)
	go func() {
		defer close(out)
		for _, i := range is {
			out <- i
		}
	}()
	return out
}
