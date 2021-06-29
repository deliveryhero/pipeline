package pipeline

import (
	"context"
	"time"

	"github.com/deliveryhero/pipeline/semaphore"
)

// ProcessBatch collects up to maxSize elements over maxDuration and processes them together as a slice of `interface{}`s.
// It passed an []interface{} to the `Processor.Process` method and expects a []interface{} back.
// It passes []interface{} batches of inputs to the `Processor.Cancel` method.
// If the receiver is backed up, ProcessBatch can holds up to 2x maxSize.
func ProcessBatch(
	ctx context.Context,
	maxSize int,
	maxDuration time.Duration,
	processor Processor,
	in <-chan interface{},
) <-chan interface{} {
	out := make(chan interface{})
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
func ProcessBatchConcurrently(
	ctx context.Context,
	concurrently,
	maxSize int,
	maxDuration time.Duration,
	processor Processor,
	in <-chan interface{},
) <-chan interface{} {
	// Create the out chan
	out := make(chan interface{})
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
func processOneBatch(
	ctx context.Context,
	maxSize int,
	maxDuration time.Duration,
	processor Processor,
	in <-chan interface{},
	out chan<- interface{},
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
			for _, result := range results.([]interface{}) {
				out <- result
			}
		}
	}
	return open
}
