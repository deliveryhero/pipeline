package pipeline_test

import (
	"context"
	"log"
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
