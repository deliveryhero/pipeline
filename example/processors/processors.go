// processors are a bunch of simple processors used in the examples
package processors

import (
	"context"
	"errors"
	"log"
	"time"
)

// Miltiplier is a simple processor that multiplies each integer it receives by some Factor
type Multiplier struct {
	// Factor will change the amount each number is multiplied by
	Factor int
}

// Process multiplies a number by factor
func (m *Multiplier) Process(_ context.Context, in interface{}) (interface{}, error) {
	return in.(int) * m.Factor, nil
}

// Cancel is called when the context is canceled
func (m *Multiplier) Cancel(i interface{}, err error) {
	log.Printf("error: could not multiply %d, %s\n", i, err)
}

// BatchMultiplier is a simple batch processor that multiplies each `[]int` it receives together
type BatchMultiplier struct{}

// Process a slice of numbers together and returns a slice of numbers with the results
func (m *BatchMultiplier) Process(_ context.Context, ins interface{}) (interface{}, error) {
	result := 1
	for _, in := range ins.([]interface{}) {
		result *= in.(int)
	}
	return []interface{}{result}, nil
}

// Cancel is called when the context is canceled
func (m *BatchMultiplier) Cancel(i interface{}, err error) {
	log.Printf("error: could not multiply %+v, %s\n", i, err)
}

// Waiter is a Processor that waits for Duration before returning its output
type Waiter struct {
	Duration time.Duration
}

// Process waits for `Waiter.Duration` before returning the value passed in
func (w *Waiter) Process(ctx context.Context, in interface{}) (interface{}, error) {
	select {
	case <-time.After(w.Duration):
		return in, nil
	case <-ctx.Done():
		return nil, errors.New("process was canceled")
	}
}

// Cancel is called when the context is canceled
func (w *Waiter) Cancel(i interface{}, err error) {
	log.Printf("error: could not process %+v, %s\n", i, err)
}
