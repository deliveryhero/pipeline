package pipeline

import (
	"context"
	"sync"
	"time"
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
		processBatch(ctx, maxSize, maxDuration, processor, in, out)
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
			processBatch(ctx, maxSize, maxDuration, processor, in, out)
			wg.Done()
		}()
	}
	return out
}

func processBatch(
	ctx context.Context,
	maxSize int,
	maxDuration time.Duration,
	processor Processor,
	in <-chan interface{},
	out chan<- interface{},
) {
	for {
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
					continue
				}
				// Split the results back into interfaces
				for _, result := range results.([]interface{}) {
					out <- result
				}
			}
		}
		// In is closed
		if !open {
			return
		}
	}
}
