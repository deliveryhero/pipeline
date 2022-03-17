package pipeline_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deliveryhero/pipeline/v2"
)

// ExamplePipelineShutsDownOnError is an example that shows how you can shutdown a pipeline
// gracefully when it receives an error.
func ExampleProcessorShutsDownOnError() {
	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a pipeline that emits 1-10
	p := pipeline.Emit(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	// A step that will shutdown the pipeline if the number is greater than 1
	p = pipeline.Process(ctx, pipeline.NewProcessor(func(ctx context.Context, i int) (int, error) {
		// Shut down the pipeline by canceling the context
		if i != 1 {
			cancel()
			return i, fmt.Errorf("%d caused the shutdown", i)
		}
		return i, nil
	}, func(i int, err error) {
		// The cancel func is called when an error is returned by the process func or the context is canceled
		fmt.Printf("could not process %d: %s\n", i, err)
	}), p)

	// Finally, lets print the results and see what happened
	for result := range p {
		fmt.Printf("result: %d\n", result)
	}

	fmt.Println("exiting the pipeline after all data is processed")

	// Output
	// could not process 2: 2 caused the shutdown
	// result: 1
	// could not process 3: context canceled
	// could not process 4: context canceled
	// could not process 5: context canceled
	// could not process 6: context canceled
	// could not process 7: context canceled
	// could not process 8: context canceled
	// could not process 9: context canceled
	// could not process 10: context canceled
	// exiting the pipeline after all data is processed
}

// ExamplePipelineShutdownWhenInputChannelIsClosed demonstrates a pipline
// that naturally finishes its run when the input channel is closed
func ExamplePipelineShutdownWhenInputChannelIsClosed() {
	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a pipeline that emits 1-10 and then closes its output channel
	p := pipeline.Emit(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	// Multiply every number by 2
	p = pipeline.Process(ctx, pipeline.NewProcessor(func(ctx context.Context, i int) (int, error) {
		return i * 2, nil
	}, func(i int, err error) {
		fmt.Printf("could not multiply %d: %s\n", i, err)
	}), p)

	// Finally, lets print the results and see what happened
	for result := range p {
		fmt.Printf("result: %d\n", result)
	}

	fmt.Println("exiting after the input channel is closed")

	// Output
	// result: 2
	// result: 4
	// result: 6
	// result: 8
	// result: 10
	// result: 12
	// result: 14
	// result: 16
	// result: 18
	// result: 20
	// exiting after the input channel is closed
}

// ExampleLongRunningPipeline demonstrates a pipline
// that runs until the os kills it
func ExampleLongRunningPipeline() {
	// Gracefully shutdown the pipeline when the the system is shutting down
	// by canceling the context when os.Kill or os.Interrupt signal is sent
	ctx, cancel := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer cancel()

	// Create a pipeline that keeps emiting numbers sequentially until the context is cancelled
	var count int
	p := pipeline.Emitter(ctx, func() int {
		count++
		return count
	})

	// Filter out only even numbers
	p = pipeline.Process(ctx, pipeline.NewProcessor(func(ctx context.Context, i int) (int, error) {
		if i%2 == 0 {
			return i, nil
		}
		return i, fmt.Errorf("'%d' is an odd number", i)
	}, func(i int, err error) {
		fmt.Printf("error processing '%v': %s\n", i, err)
	}), p)

	// Wait a few nanoseconds an simulate the os.Interrupt signal
	go func() {
		time.Sleep(time.Millisecond / 10)
		fmt.Print("\n--- os kills the app ---\n\n")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	// Finally, lets print the results and see what happened
	for result := range p {
		fmt.Printf("result: %d\n", result)
	}

	fmt.Println("exiting after the input channel is closed")

	// Output
	// error processing '1': '1' is an odd number
	// result: 2
	//
	// --- os kills the app ---
	//
	// error processing '3': '3' is an odd number
	// error processing '4': context canceled
	// exiting after the input channel is closed
}
