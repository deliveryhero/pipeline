package pipeline_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/deliveryhero/pipeline"
)

func ExamplePartition() {
	// Create a context for the pipeline
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// Create a pipeline that emits numbers 1 once, ..., 3 three times.
	p := pipeline.Emit(1, 2, 3, 2, 3, 3)

	// Create 4 partitions, one for each number 0 -> 3
	partitions := 4
	ps := pipeline.Partition(partitions, func(i interface{}) int {
		return i.(int) % partitions
	}, p)

	// Create a batch processor that sums the contents in each batch
	newBatchSumProcessor := func(partitionID int) pipeline.Processor {
		return pipeline.NewProcessor(func(ctx context.Context, i interface{}) (interface{}, error) {
			sum := len(i.([]interface{}))
			return []interface{}{fmt.Sprintf("partition %d received %d numbers", partitionID, sum)}, nil
		}, nil)
	}

	// Sum the contents of each partition with a the BatchSumProcessor
	var outs []<-chan interface{}
	for i, partition := range ps {
		outs = append(outs, pipeline.ProcessBatch(ctx, 3, time.Millisecond, newBatchSumProcessor(i), partition))
	}

	// Print the output
	for i := range pipeline.Merge(outs...) {
		fmt.Println(i)
	}

	fmt.Println("finished")

	// Output
	// partition 2 received 2 numbers
	// partition 1 received 1 numbers
	// partition 3 received 3 numbers
	// finished
}
