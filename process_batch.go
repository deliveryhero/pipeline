package pipeline

import (
	"context"
	"time"

	"github.com/deliveryhero/pipeline/v2/semaphore"
)

// ProcessBatch collects up to maxSize elements over maxDuration and processes them together as a slice of `Input`s.
// It passed an []Output to the `Processor.Process` method and expects a []Input back.
// It passes []Input batches of inputs to the `Processor.Cancel` method.
// If the receiver is backed up, ProcessBatch can holds up to 2x maxSize.
func ProcessBatch[Input, Output any](
	ctx context.Context,
	maxSize int,
	maxDuration time.Duration,
	processor Processor[[]Input, []Output],
	in <-chan Input,
) <-chan Output {
	out := make(chan Output)
	go func() {
		for {
			if !processOneBatch(ctx, maxSize, maxDuration, processor, in, out) {
				break
			}
		}
		close(out)
	}()
	return out
}

// ProcessBatchConcurrently fans the in channel out to multiple batch Processors running concurrently,
// then it fans the out channels of the batch Processors back into a single out chan
func ProcessBatchConcurrently[Input, Output any](
	ctx context.Context,
	concurrently,
	maxSize int,
	maxDuration time.Duration,
	processor Processor[[]Input, []Output],
	in <-chan Input,
) <-chan Output {
	// Create the out chan
	out := make(chan Output)
	go func() {
		// Perform Process concurrently times
		sem := semaphore.New(concurrently)
		lctx, done := context.WithCancel(context.Background())
		for !isDone(lctx) {
			sem.Add(1)
			go func() {
				if !processOneBatch(ctx, maxSize, maxDuration, processor, in, out) {
					done()
				}
				sem.Done()
			}()
		}
		// Close the out chan after all of the Processors finish executing
		sem.Wait()
		close(out)
		done() // Satisfy go-vet
	}()
	return out
}

// isDone returns true if the context is canceled
func isDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// processOneBatch processes one batch of inputs from the in chan.
// It returns true if the in chan is still open.
func processOneBatch[Input, Output any](
	ctx context.Context,
	maxSize int,
	maxDuration time.Duration,
	processor Processor[[]Input, []Output],
	in <-chan Input,
	out chan<- Output,
) (open bool) {
	// Collect interfaces for batch processing
	is, open := collect(ctx, maxSize, maxDuration, in)
	if is != nil {
		select {
		// Cancel all inputs during shutdown
		case <-ctx.Done():
			processor.Cancel(is, ctx.Err())
		// Otherwise Process the inputs
		default:
			results, err := processor.Process(ctx, is)
			if err != nil {
				processor.Cancel(is, err)
				return open
			}
			// Split the results back into interfaces
			for _, result := range results {
				out <- result
			}
		}
	}
	return open
}
