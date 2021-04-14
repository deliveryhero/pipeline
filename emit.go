package pipeline

// Emit fans `is ...interface{}`` out to a `<-chan interface{}`
func Emit(is ...interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		for _, i := range is {
			out <- i
		}
	}()
	return out
}
