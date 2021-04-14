# pipeline

Pipeline is a go library that helps you build piplines without worrying about channel management and concurrency.
It contains common fan-in and fan-out operations as well as useful utility funcs for batch processing and scaling.

If you have another common use case you would like to see covered by this package, please [open a feature request](https://github.com/deliveryhero/pipeline/issues).

## Functions

### func [Cancel](/cancel.go#L9)

`func Cancel(ctx context.Context, cancel func(interface{}, error), in <-chan interface{}) <-chan interface{}`

Cacel passes an `interface{}` from the `in <-chan interface{}` directly to the out `<-chan interface{}` until the `Context` is canceled.
After the context is canceled, everything from `in <-chan interface{}` is sent to the `cancel` func instead with the `ctx.Err()`.

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

### func [Process](/process.go#L9)

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

	// Use the Multipleir to multiply each int by 10
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

	// Use the BatchMultipleir to multiply 2 adjacent numbers together
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
