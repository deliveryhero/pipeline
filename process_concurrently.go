package util

import (
	"context"
	"time"
)

// ProcessConcurrently fans the in channel out to multple Processors running concurrently,
// then it fans the out channles of the Processors back into a single out chan
func ProcessConcurrently(ctx context.Context, concurrently int, p Processor, in <-chan interface{}) <-chan interface{} {
	var outs []<-chan interface{}
	for i := 0; i < concurrently; i++ {
		outs = append(outs, Process(ctx, p, in))
	}
	return Merge(outs...)
}

// ProcessBatchConcurrently fans the in channel out to multple batch Processors running concurrently,
// then it fans the out channles of the batch Processors back into a single out chan
func ProcessBatchConcurrently(
	ctx context.Context,
	concurrently,
	maxSize int,
	maxDuration time.Duration,
	processor Processor,
	in <-chan interface{},
) <-chan interface{} {
	var outs []<-chan interface{}
	for i := 0; i < concurrently; i++ {
		outs = append(outs, ProcessBatch(ctx, maxSize, maxDuration, processor, in))
	}
	return Merge(outs...)
}
