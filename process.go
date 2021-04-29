package pipeline

import (
	"context"
	"sync"
)

// Process takes each input from the `in <-chan interface{}` and calls `Processor.Process` on it.
// When `Processor.Process` returns an `interface{}`, it will be sent to the output `<-chan interface{}`.
// If `Processor.Process` returns an error, `Processor.Cancel` will be called with the corresponding input and error message.
// Finally, if the `Context` is canceled, all inputs remaining in the `in <-chan interface{}` will go directly to `Processor.Cancel`.
func Process(ctx context.Context, processor Processor, in <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		process(ctx, processor, in, out)
		close(out)
	}()
	return out
}

// ProcessConcurrently fans the in channel out to multiple Processors running concurrently,
// then it fans the out channels of the Processors back into a single out chan
func ProcessConcurrently(ctx context.Context, concurrently int, p Processor, in <-chan interface{}) <-chan interface{} {
	// Create the out chan
	out := make(chan interface{})
	// Close the out chan after all of the Processors finish executing
	var wg sync.WaitGroup
	wg.Add(concurrently)
	go func() {
		wg.Wait()
		close(out)
	}()
	// Perform Process concurrently times
	for i := 0; i < concurrently; i++ {
		go func() {
			process(ctx, p, in, out)
			wg.Done()
		}()
	}
	return out
}

func process(
	ctx context.Context,
	processor Processor,
	in <-chan interface{},
	out chan<- interface{},
) {
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
}
