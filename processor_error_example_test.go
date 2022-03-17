package pipeline_test

import (
	"context"
	"fmt"

	pipeline "github.com/deliveryhero/pipeline/v2-beta"
)

// ExamplePipelineShutsDownOnError is an example that shows how you can shutdown a pipeline
// when it receives an error.
func ExampleProcessorShutsDownOnError() {
	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a pipeline that emits 1-10
	p := pipeline.Emit[interface{}](1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	// A step that will shutdown the pipeline if it fails
	p = pipeline.Process(ctx, pipeline.NewProcessor[interface{}, interface{}](func(ctx context.Context, i interface{}) (interface{}, error) {
		v, ok := i.(int)
		if v != 1 || !ok {
			// Shut down the pipeline by canceling the context
			cancel()
			return nil, fmt.Errorf("%d caused the shutdown", v)
		}
		return v, nil
	}, func(i interface{}, err error) {
		// The cancel func is called when the context is canceled
		// as well as when we return an error from the process func
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
