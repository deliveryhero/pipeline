package pipeline_test

import (
	"context"
	"log"
	"time"

	"github.com/deliveryhero/pipeline"
)

// Miltiplier is a simple processor that multiplies integers by some Factor
type Multiplier struct {
	// Factor will change the amount each number is multiplied by
	Factor int
}

// Process multiplies a number by factor
func (m *Multiplier) Process(_ context.Context, in interface{}) (interface{}, error) {
	return in.(int) * m.Factor, nil
}

// Cancel is called when the context is cancelled
func (m *Multiplier) Cancel(i interface{}, err error) {
	log.Printf("error: could not multiply %d, %s\n", i, err)
}

func ExampleProcessor() {
	// m is a processor that multiplies numbers by 10
	m := &Multiplier{
		Factor: 10,
	}
	log.Println(m.Process(context.Background(), 1))

	// Output
	// 10
}

func ExampleProcess() {
	// Create a context that times out after 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a pipeline that emits 1-6 at a rate of one int per second
	p := pipeline.Delay(ctx, time.Second, pipeline.Emit(1, 2, 3, 4, 5, 6))

	// Use the Multipleir to multiply each int by 10
	p = pipeline.Process(ctx, &Multiplier{
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
