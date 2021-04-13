package util

// Emit fans is out into a channel
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
