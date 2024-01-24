package pipeline_test

import (
	"context"
	"fmt"
	"time"

	"github.com/deliveryhero/pipeline/v2"
)

func ExampleProcessBatch() {
	// Create a context that times out after 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a pipeline that emits 1-6 at a rate of one int per second
	p := pipeline.Delay(ctx, time.Second, pipeline.Emit(1, 2, 3, 4, 5, 6))

	// Multiply every 2 adjacent numbers together
	p = pipeline.ProcessBatch(ctx, 2, time.Minute, pipeline.NewProcessor(func(ctx context.Context, is []int) ([]int, error) {
		o := 1
		for _, i := range is {
			o *= i
		}
		return []int{o}, nil
	}, func(is []int, err error) {
		fmt.Printf("error: could not multiply %v, %s\n", is, err)
	}), p)

	// Finally, lets print the results and see what happened
	for result := range p {
		fmt.Printf("result: %d\n", result)
	}

	// Output:
	// result: 2
	// result: 12
	// error: could not multiply [5 6], context deadline exceeded
}

func ExampleProcessBatchConcurrently() {
	// Create a context that times out after 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a pipeline that emits 1-9
	p := pipeline.Emit(1, 2, 3, 4, 5, 6, 7, 8, 9)

	// Add a 1 second delay to each number
	p = pipeline.Delay(ctx, time.Second, p)

	// Group two inputs at a time
	p = pipeline.ProcessBatchConcurrently(ctx, 2, 2, time.Minute, pipeline.NewProcessor(func(ctx context.Context, ins []int) ([]int, error) {
		return ins, nil
	}, func(i []int, err error) {
		fmt.Printf("error: could not process %v, %s\n", i, err)
	}), p)

	// Finally, lets print the results and see what happened
	for result := range p {
		fmt.Printf("result: %d\n", result)
	}

	// Example Output
	// result: 1
	// result: 2
	// result: 3
	// result: 5
	// error: could not process [7 8], context deadline exceeded
	// error: could not process [4 6], context deadline exceeded
	// error: could not process [9], context deadline exceeded
}
