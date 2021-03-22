package util

import "context"

// Processor processes an input and reurns an output
type Processor interface {
	// Process processes an input and reurns an output
	Process(i interface{}) (interface{}, error)

	// Cancel is called if the context is canceled while the input is processing
	Cancel(i interface{}, err error)
}

// Process processes each input and returns a cooresponding output
func Process(ctx context.Context, p Processor, in <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	// process calls Processor.Process asynchronously
	process := func(i interface{}) <-chan interface{} {
		out := make(chan interface{})
		go func() {
			defer close(out)
			result, err := p.Process(i)
			if err != nil {
				p.Cancel(i, err)
				return
			}
			out <- result
		}()
		return out
	}
	go func() {
		defer close(out)
		// Start processing inputs until in closes
		for i := range in {
			select {
			// Process one input
			case o, open := <-process(i):
				if open {
					out <- o
				}
			// Cancel all inputs if the context is closed
			case <-ctx.Done():
				p.Cancel(i, ctx.Err())
			}
		}
	}()
	return out
}
