package pipeline_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/deliveryhero/pipeline"
	"github.com/deliveryhero/pipeline/example/processors"
)

func ExampleProcessBatch() {
	// Create a context that times out after 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a pipeline that emits 1-6 at a rate of one int per second
	p := pipeline.Delay(ctx, time.Second, pipeline.Emit(1, 2, 3, 4, 5, 6))

	// Use the BatchMultiplier to multiply 2 adjacent numbers together
	p = pipeline.ProcessBatch(ctx, 2, time.Minute, &processors.BatchMultiplier{}, p)

	// Finally, lets print the results and see what happened
	for result := range p {
		log.Printf("result: %d\n", result)
	}

	// Output
	// result: 2
	// result: 12
	// error: could not multiply [5], context deadline exceeded
	// error: could not multiply [6], context deadline exceeded
}

func ExampleProcessBatchConcurrently() {
	// Create a context that times out after 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a pipeline that emits 1-9
	p := pipeline.Emit(1, 2, 3, 4, 5, 6, 7, 8, 9)

	// Wait 4 seconds to pass 2 numbers through the pipe
	// * 2 concurrent Processors
	p = pipeline.ProcessBatchConcurrently(ctx, 2, 2, time.Minute, &processors.Waiter{
		Duration: 4 * time.Second,
	}, p)

	// Finally, lets print the results and see what happened
	for result := range p {
		log.Printf("result: %d\n", result)
	}

	// Output
	// result: 3
	// result: 4
	// result: 1
	// result: 2
	// error: could not process [5 6], process was canceled
	// error: could not process [7 8], process was canceled
	// error: could not process [9], context deadline exceeded
}

func ExampleProcessBatchPartitionsConcurrently() {
	// Create a context for the pipeline
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// Create a pipeline that emits numbers 1 once, ..., 3 three times.
	// Add a 0 to demonstrate error handling
	p := pipeline.Emit(0, 1, 2, 3, 2, 3, 3)

	// Create a partition func that will route each number into a seprate partition
	partitions := 3
	partition := func(i interface{}) int {
		return i.(int) % partitions
	}

	// Create a batch processor that sums the contents of each batch
	processor := pipeline.NewProcessor(func(ctx context.Context, i interface{}) (interface{}, error) {
		is := i.([]interface{})
		// If  each partition to contain only one number
		prev := is[0].(int)
		for _, i := range is {
			n := i.(int)
			if prev != n {
				return nil, fmt.Errorf("the batch contains more than one number: %d != %d", prev, n)
			}
			prev = n
		}
		return []interface{}{fmt.Sprintf("partition %d received %d numbers", prev, len(is))}, nil
	}, func(i interface{}, err error) {
		fmt.Printf("error: %s\n", err)
	})

	// Partition the pipeline into 3 batches (indexed 0 -> 2)
	p = pipeline.ProcessBatchPartitionsConcurrently(ctx, partitions, 5, time.Second, partition, processor, p)

	// Print the output
	for i := range p {
		fmt.Println(i)
	}

	fmt.Println("finished")

	// Output
	// error: the batch contains more than one number: 0 != 3
	// partition 2 received 2 numbers
	// partition 1 received 1 numbers
	// finished
}
