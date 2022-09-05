package pipeline_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/deliveryhero/pipeline/v2"
)

func ExampleProcess() {
	// Create a context that times out after 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a pipeline that emits 1-6 at a rate of one int per second
	p := pipeline.Delay(ctx, time.Second, pipeline.Emit(1, 2, 3, 4, 5, 6))

	// Multiply each number by 10
	p = pipeline.Process(ctx, pipeline.NewProcessor(func(ctx context.Context, in int) (int, error) {
		return in * 10, nil
	}, func(i int, err error) {
		fmt.Printf("error: could not multiply %v, %s\n", i, err)
	}), p)

	// Finally, lets print the results and see what happened
	for result := range p {
		fmt.Printf("result: %d\n", result)
	}

	// Output:
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

	// Add a two second delay to each number
	p = pipeline.Delay(ctx, 2*time.Second, p)

	// Add two concurrent processors that pass input numbers to the output
	p = pipeline.ProcessConcurrently(ctx, 2, pipeline.NewProcessor(func(ctx context.Context, in int) (int, error) {
		return in, nil
	}, func(i int, err error) {
		fmt.Printf("error: could not process %v, %s\n", i, err)
	}), p)

	// Finally, lets print the results and see what happened
	for result := range p {
		log.Printf("result: %d\n", result)
	}

	// Output:
	// result: 2
	// result: 1
	// result: 4
	// result: 3
	// error: could not process 6, process was canceled
	// error: could not process 5, process was canceled
	// error: could not process 7, context deadline exceeded
}
