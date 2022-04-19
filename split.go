package pipeline

// Split takes an interface from Collect and splits it back out into individual elements
func Split[Item any](in <-chan []Item) <-chan Item {
	out := make(chan Item)
	go func() {
		defer close(out)
		for is := range in {
			for _, i := range is {
				out <- i
			}
		}
	}()
	return out
}
