package pipeline

import (
	"context"

	"github.com/deliveryhero/pipeline/semaphore"
)

// Process takes each input from the `in <-chan interface{}` and calls `Processor.Process` on it.
// When `Processor.Process` returns an `interface{}`, it will be sent to the output `<-chan interface{}`.
// If `Processor.Process` returns an error, `Processor.Cancel` will be called with the corresponding input and error message.
// Finally, if the `Context` is canceled, all inputs remaining in the `in <-chan interface{}` will go directly to `Processor.Cancel`.
func Process(ctx context.Context, processor Processor, in <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		for i := range in {
			process(ctx, processor, i, out)
		}
		close(out)
	}()
	return out
}

// ProcessConcurrently fans the in channel out to multiple Processors running concurrently,
// then it fans the out channels of the Processors back into a single out chan
func ProcessConcurrently(ctx context.Context, concurrently int, p Processor, in <-chan interface{}) <-chan interface{} {
	// Create the out chan
	out := make(chan interface{})
	go func() {
		// Perform Process concurrently times
		sem := semaphore.New(concurrently)
		for i := range in {
			sem.Add(1)
			go func(i interface{}) {
				process(ctx, p, i, out)
				sem.Done()
			}(i)
		}
		// Close the out chan after all of the Processors finish executing
		sem.Wait()
		close(out)
	}()
	return out
}

func process(
	ctx context.Context,
	processor Processor,
	i interface{},
	out chan<- interface{},
) {
	select {
	// When the context is canceled, Cancel all inputs
	case <-ctx.Done():
		processor.Cancel(i, ctx.Err())
	// Otherwise, Process all inputs
	default:
		result, err := processor.Process(ctx, i)
		if err != nil {
			processor.Cancel(i, err)
			return
		}
		out <- result
	}
}
