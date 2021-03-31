package util

import (
	"context"
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
	// 3. Split the processed slices of inputs back out into individial outputs
	return Split(
		// 2. Process the slices of inputs
		Process(
			ctx,
			processor,
			// 1. Collect slices of inputs from in
			Collect(
				ctx,
				maxSize,
				maxDuration,
				in,
			),
		),
	)
}
