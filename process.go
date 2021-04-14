package pipeline

import "context"

// Processor represents a blocking operation in a pipeline. Implementing `Processor` will allow you to add
// business logic to your pipelines without directly managing channels. This simplifies your unit tests
// and eliminates channel management related bugs.
type Processor interface {
	// Process processes an input and reurns an output or an error, if the output could not be processed.
	// When the context is canceled, process should stop all blocking operations and return the `Context.Err()`.
	Process(ctx context.Context, i interface{}) (interface{}, error)

	// Cancel is called if process returns an error or if the context is canceled while there are still items in the `in <-chan interface{}`.
	Cancel(i interface{}, err error)
}

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
