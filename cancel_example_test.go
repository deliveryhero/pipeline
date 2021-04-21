package pipeline_test

import (
	"context"
	"log"
	"time"

	"github.com/deliveryhero/pipeline"
)

func ExampleCancel() {
	// Create a context that lasts for 1 second
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Create a basic pipeline that emits one int every 250ms
	p := pipeline.Delay(ctx, time.Second/4,
		pipeline.Emit(1, 2, 3, 4, 5),
	)

	// If the context is canceled, pass the ints to the cancel func for teardown
	p = pipeline.Cancel(ctx, func(i interface{}, err error) {
		log.Printf("%+v could not be processed, %s", i, err)
	}, p)

	// Otherwise, process the inputs
	for out := range p {
		log.Printf("process: %+v", out)
	}

	// Output
	// process: 1
	// process: 2
	// process: 3
	// process: 4
	// 5 could not be processed, context deadline exceeded
}
