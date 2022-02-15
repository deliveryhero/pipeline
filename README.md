# pipeline

[![GitHub Workflow Status](https://github.com/deliveryhero/pipeline/actions/workflows/ci.yml/badge.svg?branch=partition-func)](https://github.com/deliveryhero/pipeline/actions/workflows/ci.yml?query=branch:partition-func)
[![codecov](https://codecov.io/gh/deliveryhero/pipeline/branch/partition-func/graph/badge.svg)](https://codecov.io/gh/deliveryhero/pipeline)
[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/deliveryhero/pipeline)
[![Go Report Card](https://goreportcard.com/badge/github.com/deliveryhero/pipeline)](https://goreportcard.com/report/github.com/deliveryhero/pipeline)

Pipeline is a go library that helps you build pipelines without worrying about channel management and concurrency.
It contains common fan-in and fan-out operations as well as useful utility funcs for batch processing and scaling.

If you have another common use case you would like to see covered by this package, please [open a feature request](https://github.com/deliveryhero/pipeline/issues).

## Functions

### func [Buffer](/buffer.go#L5)

`func Buffer(size int, in <-chan interface{}) <-chan interface{}`

Buffer creates a buffered channel that will close after the input
is closed and the buffer is fully drained

### func [Cancel](/cancel.go#L9)

`func Cancel(ctx context.Context, cancel func(interface{}, error), in <-chan interface{}) <-chan interface{}`

Cancel passes an `interface{}` from the `in <-chan interface{}` directly to the out `<-chan interface{}` until the `Context` is canceled.
After the context is canceled, everything from `in <-chan interface{}` is sent to the `cancel` func instead with the `ctx.Err()`.

```golang
package main

import (
	"context"
	"github.com/deliveryhero/pipeline"
	"log"
	"time"
)

func main() {
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

```

### func [Collect](/collect.go#L13)

`func Collect(ctx context.Context, maxSize int, maxDuration time.Duration, in <-chan interface{}) <-chan interface{}`

Collect collects `interface{}`s from its in channel and returns `[]interface{}` from its out channel.
It will collect up to `maxSize` inputs from the `in <-chan interface{}` over up to `maxDuration` before returning them as `[]interface{}`.
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

```golang
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/deliveryhero/pipeline"
	"github.com/deliveryhero/pipeline/example/db"
)

// SearchResults returns many types of search results at once
type SearchResults struct {
	Advertisements []db.Result `json:"advertisements"`
	Images         []db.Result `json:"images"`
	Products       []db.Result `json:"products"`
	Websites       []db.Result `json:"websites"`
}

func main() {
	r := http.NewServeMux()

	// `GET /search?q=<query>` is an endpoint that merges concurrently fetched
	// search results into a single search response using `pipeline.Merge`
	r.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if len(query) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// If the request times out, or we receive an error from our `db`
		// the context will stop all pending db queries for this request
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		// Fetch all of the different search results concurrently
		var results SearchResults
		for err := range pipeline.Merge(
			db.GetAdvertisements(ctx, query, &results.Advertisements),
			db.GetImages(ctx, query, &results.Images),
			db.GetProducts(ctx, query, &results.Products),
			db.GetWebsites(ctx, query, &results.Websites),
		) {
			// Stop all pending db requests if theres an error
			if err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// Return the search results
		if bs, err := json.Marshal(&results); err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else if _, err := w.Write(bs); err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
}

```

### func [Partition](/partition.go#L5)

`func Partition(partitions int, partition func(interface{}) int, in <-chan interface{}) []<-chan interface{}`

Partition fans a single input channel into multiple output channels based on the partition func.
Warning: If `partition(interface{}) int` returns an int >= `partitions`, it will cause a panic.

```golang
package main

import (
	"context"
	"fmt"
	"github.com/deliveryhero/pipeline"
	"os"
	"os/signal"
	"time"
)

func main() {
	// Create a context for the pipeline
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// Create a pipeline that emits numbers 1 once, ..., 3 three times.
	p := pipeline.Emit(1, 2, 3, 2, 3, 3)

	// Create 4 partitions, one for each number 0 -> 3
	partitions := 4
	ps := pipeline.Partition(partitions, func(i interface{}) int {
		return i.(int) % partitions
	}, p)

	// Create a batch processor that sums the contents in each batch
	newBatchSumProcessor := func(partitionID int) pipeline.Processor {
		return pipeline.NewProcessor(func(ctx context.Context, i interface{}) (interface{}, error) {
			sum := len(i.([]interface{}))
			return []interface{}{fmt.Sprintf("partition %d received %d numbers", partitionID, sum)}, nil
		}, nil)
	}

	// Sum the contents of each partition with a the BatchSumProcessor
	var outs []<-chan interface{}
	for i, partition := range ps {
		outs = append(outs, pipeline.ProcessBatch(ctx, 3, time.Millisecond, newBatchSumProcessor(i), partition))
	}

	// Print the output
	for i := range pipeline.Merge(outs...) {
		fmt.Println(i)
	}

	fmt.Println("finished")

	// Output
	// partition 2 received 2 numbers
	// partition 1 received 1 numbers
	// partition 3 received 3 numbers
	// finished
}

```

### func [Process](/process.go#L13)

`func Process(ctx context.Context, processor Processor, in <-chan interface{}) <-chan interface{}`

Process takes each input from the `in <-chan interface{}` and calls `Processor.Process` on it.
When `Processor.Process` returns an `interface{}`, it will be sent to the output `<-chan interface{}`.
If `Processor.Process` returns an error, `Processor.Cancel` will be called with the corresponding input and error message.
Finally, if the `Context` is canceled, all inputs remaining in the `in <-chan interface{}` will go directly to `Processor.Cancel`.

```golang
package main

import (
	"context"
	"github.com/deliveryhero/pipeline"
	"github.com/deliveryhero/pipeline/example/processors"
	"log"
	"time"
)

func main() {
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

```

### func [ProcessBatch](/process_batch.go#L15)

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

```golang
package main

import (
	"context"
	"github.com/deliveryhero/pipeline"
	"github.com/deliveryhero/pipeline/example/processors"
	"log"
	"time"
)

func main() {
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

```

### func [ProcessBatchConcurrently](/process_batch.go#L36)

`func ProcessBatchConcurrently(
    ctx context.Context,
    concurrently,
    maxSize int,
    maxDuration time.Duration,
    processor Processor,
    in <-chan interface{},
) <-chan interface{}`

ProcessBatchConcurrently fans the in channel out to multiple batch Processors running concurrently,
then it fans the out channels of the batch Processors back into a single out chan

```golang
package main

import (
	"context"
	"github.com/deliveryhero/pipeline"
	"github.com/deliveryhero/pipeline/example/processors"
	"log"
	"time"
)

func main() {
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

```

### func [ProcessBatchPartitionsConcurrently](/process_batch.go#L72)

`func ProcessBatchPartitionsConcurrently(
    ctx context.Context,
    partitions,
    maxSize int,
    maxDuration time.Duration,
    partition func(i interface{}) int,
    processor Processor,
    in <-chan interface{},
) <-chan interface{}`

ProcessBatchPartitionsConcurrently is similar to `ProcessBatchConcurrently`, except
it uses a custom partition function to determine which batch `Processor` each input gets sent to.
This is useful for preventing locking issues while batch proccessing data concurrently.
If you are not encountering locking issues, it is slightly more efficient to use `ProcessBatchConcurrently` instead.
Warning: If `partition(interface{}) int` returns an int >= `partitions`, it will cause a panic.

```golang
package main

import (
	"context"
	"fmt"
	"github.com/deliveryhero/pipeline"
	"os"
	"os/signal"
	"time"
)

func main() {
	// Create a context for the pipeline
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// Create a pipeline that emits numbers 1 once, ..., 3 three times.
	// Add a 0 to demonstrate error handling
	p := pipeline.Emit(0, 1, 2, 3, 2, 3, 3)

	// Create a partition func that will route each number into a seprate partition
	partitions := 3
	partition := func(i interface{}) int {
		return i.(int) % partitions
	}

	// Create a batch processor that sums the contents of each batch
	processor := pipeline.NewProcessor(func(ctx context.Context, i interface{}) (interface{}, error) {
		is := i.([]interface{})
		// If  each partition to contain only one number
		prev := is[0].(int)
		for _, i := range is {
			n := i.(int)
			if prev != n {
				return nil, fmt.Errorf("the batch contains more than one number: %d != %d", prev, n)
			}
			prev = n
		}
		return []interface{}{fmt.Sprintf("partition %d received %d numbers", prev, len(is))}, nil
	}, func(i interface{}, err error) {
		fmt.Printf("error: %s\n", err)
	})

	// Partition the pipeline into 3 batches (indexed 0 -> 2)
	p = pipeline.ProcessBatchPartitionsConcurrently(ctx, partitions, 5, time.Second, partition, processor, p)

	// Print the output
	for i := range p {
		fmt.Println(i)
	}

	fmt.Println("finished")

	// Output
	// error: the batch contains more than one number: 0 != 3
	// partition 2 received 2 numbers
	// partition 1 received 1 numbers
	// finished
}

```

### func [ProcessConcurrently](/process.go#L26)

`func ProcessConcurrently(ctx context.Context, concurrently int, p Processor, in <-chan interface{}) <-chan interface{}`

ProcessConcurrently fans the in channel out to multiple Processors running concurrently,
then it fans the out channels of the Processors back into a single out chan

```golang
package main

import (
	"context"
	"github.com/deliveryhero/pipeline"
	"github.com/deliveryhero/pipeline/example/processors"
	"log"
	"time"
)

func main() {
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

```

### func [Split](/split.go#L5)

`func Split(in <-chan interface{}) <-chan interface{}`

Split takes an interface from Collect and splits it back out into individual elements
Useful for batch processing pipelines (`input chan -> Collect -> Process -> Split -> Cancel -> output chan`).

## Types

### type [Processor](/processor.go#L8)

`type Processor interface { ... }`

Processor represents a blocking operation in a pipeline. Implementing `Processor` will allow you to add
business logic to your pipelines without directly managing channels. This simplifies your unit tests
and eliminates channel management related bugs.

## Sub Packages

* [semaphore](./semaphore): package semaphore is like a sync.WaitGroup with an upper limit.

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
