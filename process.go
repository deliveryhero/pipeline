package pipeline

import (
	"context"

	"github.com/deliveryhero/pipeline/v2/semaphore"
)

// Process takes each input from the `in <-chan Input` and calls `Processor.Process` on it.
// When `Processor.Process` returns an `Output`, it will be sent to the output `<-chan Output`.
// If `Processor.Process` returns an error, `Processor.Cancel` will be called with the corresponding input and error message.
// Finally, if the `Context` is canceled, all inputs remaining in the `in <-chan Input` will go directly to `Processor.Cancel`.
func Process[Input, Output any](ctx context.Context, processor Processor[Input, Output], in <-chan Input) <-chan Output {
	out := make(chan Output)
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
func ProcessConcurrently[Input, Output any](ctx context.Context, concurrently int, p Processor[Input, Output], in <-chan Input) <-chan Output {
	// Create the out chan
	out := make(chan Output)
	go func() {
		// Perform Process concurrently times
		sem := semaphore.New(concurrently)
		for i := range in {
			sem.Add(1)
			go func(i Input) {
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

func process[A, B any](
	ctx context.Context,
	processor Processor[A, B],
	i A,
	out chan<- B,
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
