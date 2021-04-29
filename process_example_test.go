package pipeline_test

import (
	"context"
	"log"
	"time"

	"github.com/deliveryhero/pipeline"
	"github.com/deliveryhero/pipeline/example/processors"
)

func ExampleProcess() {
	// Create a context that times out after 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a pipeline that emits 1-6 at a rate of one int per second
	p := pipeline.Delay(ctx, time.Second, pipeline.Emit(1, 2, 3, 4, 5, 6))

	// Use the Multiplier to multiply each int by 10
	p = pipeline.Process(ctx, &processors.Multiplier{
		Factor: 10,
	}, p)

	// Finally, lets print the results and see what happened
	for result := range p {
		log.Printf("result: %d\n", result)
	}

	// Output
	// result: 10
	// result: 20
	// result: 30
	// result: 40
	// result: 50
	// error: could not multiply 6, context deadline exceeded
}

func ExampleProcessConcurrently() {
	// Create a context that times out after 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a pipeline that emits 1-7
	p := pipeline.Emit(1, 2, 3, 4, 5, 6, 7)

	// Wait 2 seconds to pass each number through the pipe
	// * 2 concurrent Processors
	p = pipeline.ProcessConcurrently(ctx, 2, &processors.Waiter{
		Duration: 2 * time.Second,
	}, p)

	// Finally, lets print the results and see what happened
	for result := range p {
		log.Printf("result: %d\n", result)
	}

	// Output
	// result: 2
	// result: 1
	// result: 4
	// result: 3
	// error: could not process 6, process was canceled
	// error: could not process 5, process was canceled
	// error: could not process 7, context deadline exceeded
}
