package pipeline

import (
	"context"
	"time"
)

// ProcessConcurrently fans the in channel out to multiple Processors running concurrently,
// then it fans the out channels of the Processors back into a single out chan
func ProcessConcurrently(ctx context.Context, concurrently int, p Processor, in <-chan interface{}) <-chan interface{} {
	var outs []<-chan interface{}
	for i := 0; i < concurrently; i++ {
		outs = append(outs, Process(ctx, p, in))
	}
	return Merge(outs...)
}

// ProcessBatchConcurrently fans the in channel out to multiple batch Processors running concurrently,
// then it fans the out channels of the batch Processors back into a single out chan
func ProcessBatchConcurrently(
	ctx context.Context,
	concurrently,
	maxSize int,
	maxDuration time.Duration,
	p Processor,
	in <-chan interface{},
) <-chan interface{} {
	var outs []<-chan interface{}
	for i := 0; i < concurrently; i++ {
		outs = append(outs, ProcessBatch(ctx, maxSize, maxDuration, p, in))
	}
	return Merge(outs...)
}
