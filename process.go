package pipeline

import "context"

// Process takes each input from the `in <-chan interface{}` and calls `Processor.Process` on it.
// When `Processor.Process` returns an `interface{}`, it will be sent to the output `<-chan interface{}`.
// If `Processor.Process` returns an error, `Processor.Cancel` will be called with the cooresponding input and error message.
// Finally, if the `Context` is cancelled, all inputs remaining in the `in <-chan interface{}` will go directly to `Processor.Cancel`.
func Process(ctx context.Context, processor Processor, in <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		// Start processing inputs until in closes
		for i := range in {
			select {
			// When the context is canceled, Cancel all inputs
			case <-ctx.Done():
				processor.Cancel(i, ctx.Err())
			// Otherwise, Process all inputs
			default:
				result, err := processor.Process(ctx, i)
				if err != nil {
					processor.Cancel(i, err)
					continue
				}
				out <- result
			}
		}
	}()
	return out
}
