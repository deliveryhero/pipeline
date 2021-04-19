package pipeline

import "context"

// Processor represents a blocking operation in a pipeline. Implementing `Processor` will allow you to add
// business logic to your pipelines without directly managing channels. This simplifies your unit tests
// and eliminates channel management related bugs.
type Processor interface {
	// Process processes an input and returns an output or an error, if the output could not be processed.
	// When the context is canceled, process should stop all blocking operations and return the `Context.Err()`.
	Process(ctx context.Context, i interface{}) (interface{}, error)

	// Cancel is called if process returns an error or if the context is canceled while there are still items in the `in <-chan interface{}`.
	Cancel(i interface{}, err error)
}

// NewProcessor creates a process and cancel func
func NewProcessor(
	process func(ctx context.Context, i interface{}) (interface{}, error),
	cancel func(i interface{}, err error),
) Processor {
	return &processor{process, cancel}
}

// processor implements Processor
type processor struct {
	process func(ctx context.Context, i interface{}) (interface{}, error)
	cancel  func(i interface{}, err error)
}

func (p *processor) Process(ctx context.Context, i interface{}) (interface{}, error) {
	return p.process(ctx, i)
}

func (p *processor) Cancel(i interface{}, err error) {
	p.cancel(i, err)
}
