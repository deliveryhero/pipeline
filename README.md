# pipeline

[![codecov](https://codecov.io/gh/deliveryhero/pipeline/branch/master/graph/badge.svg)](https://codecov.io/gh/deliveryhero/pipeline)
[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/deliveryhero/pipeline)
[![Go Report Card](https://goreportcard.com/badge/github.com/deliveryhero/pipeline)](https://goreportcard.com/report/github.com/deliveryhero/pipeline)

Pipeline is a go library for helping you build piplines without worrying about channel management and concurrency.
It contains common fan-in and fan-out operations as well as useful `func`s for batch processing and concurrency.
If you have another common use case you would like to see covered, please [open an issue](https://github.com/deliveryhero/pipeline/issues).

## Functions

### func [Cancel](/cancel.go#L7)

`func Cancel(ctx context.Context, cancel func(interface{}, error), in <-chan interface{}) <-chan interface{}`

Cacel passes `interface{}`s from the in chan to the out chan until the context is canceled.
When the context is canceled, everything from in is intercepted by the `cancel` func instead of being passed to the out chan.

### func [Collect](/collect.go#L13)

`func Collect(ctx context.Context, maxSize int, maxDuration time.Duration, in <-chan interface{}) <-chan interface{}`

Collect collects `interface{}`s from its in channel and returns `[]interface{}` from its out channel.
It will collet up to `maxSize` inputs from the `in <-chan interface{}` over up to `maxDuration` before returning them as `[]interface{}`.
That means when `maxSize` is reached before `maxDuration`, `[maxSize]interface{}` will be passed to the out channel.
But if `maxDuration` is reached before `maxSize` inputs are collected, `[< maxSize]interface{}` will be passed to the out channel.
When the `context` is canceled, everything in the buffer will be flushed to the out channel.

### func [Delay](/delay.go#L10)

`func Delay(ctx context.Context, duration time.Duration, in <-chan interface{}) <-chan interface{}`

Delay delays reading each input by `duration`.
If the context is canceled, the delay will not be applied.

### func [Emit](/emit.go#L4)

`func Emit(is ...interface{}) <-chan interface{}`

Emit fans `is ...interface{}`` out to a `<-chan interface{}`

### func [Merge](/merge.go#L6)

`func Merge(ins ...<-chan interface{}) <-chan interface{}`

Merge fans multiple channels in to a single channel

### func [Process](/process.go#L21)

`func Process(ctx context.Context, processor Processor, in <-chan interface{}) <-chan interface{}`

Process takes each input from the `in <-chan interface{}` and calls `Processor.Process` on it.
When `Processor.Process` returns an `interface{}`, it will be sent to the output `<-chan interface{}`.
If `Processor.Process` returns an error, `Processor.Cancel` will be called with the cooresponding input and error message.
Finally, if the `Context` is cancelled, all inputs remaining in the `in <-chan interface{}` will go directly to `Processor.Cancel`.

```golang
package main

import (
	"context"
	"github.com/deliveryhero/pipeline"
	"log"
	"time"
)

// Miltiplier is a simple processor that multiplies integers by some Factor
type Multiplier struct {
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

func main() {
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

```

### func [ProcessBatch](/process_batch.go#L12)

`func ProcessBatch(
    ctx context.Context,
    maxSize int,
    maxDuration time.Duration,
    processor Processor,
    in <-chan interface{},
) <-chan interface{}`

ProcessBatch collects up to maxSize elements over maxDuration and processes them together as a slice of `interface{}`s.
It passed an []interface{} to the `Processor.Process` method and expects a []interface{} back.
It passes []interface{} batches of inputs to the `Processor.Cancel` method.
If the receiver is backed up, ProcessBatch can holds up to 2x maxSize.

### func [ProcessBatchConcurrently](/process_concurrently.go#L20)

`func ProcessBatchConcurrently(
    ctx context.Context,
    concurrently,
    maxSize int,
    maxDuration time.Duration,
    p Processor,
    in <-chan interface{},
) <-chan interface{}`

ProcessBatchConcurrently fans the in channel out to multple batch Processors running concurrently,
then it fans the out channles of the batch Processors back into a single out chan

### func [ProcessConcurrently](/process_concurrently.go#L10)

`func ProcessConcurrently(ctx context.Context, concurrently int, p Processor, in <-chan interface{}) <-chan interface{}`

ProcessConcurrently fans the in channel out to multple Processors running concurrently,
then it fans the out channles of the Processors back into a single out chan

### func [Split](/split.go#L5)

`func Split(in <-chan interface{}) <-chan interface{}`

Split takes an interface from Collect and splits it back out into individual elements
Usefull for batch processing pipelines (`intput chan -> Collect -> Process -> Split -> Cancel -> output chan`).

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
