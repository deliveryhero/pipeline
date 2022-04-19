package pipeline

import "context"

// Processor represents a blocking operation in a pipeline. Implementing `Processor` will allow you to add
// business logic to your pipelines without directly managing channels. This simplifies your unit tests
// and eliminates channel management related bugs.
type Processor[Input, Output any] interface {
	// Process processes an input and returns an output or an error, if the output could not be processed.
	// When the context is canceled, process should stop all blocking operations and return the `Context.Err()`.
	Process(ctx context.Context, i Input) (Output, error)

	// Cancel is called if process returns an error or if the context is canceled while there are still items in the `in <-chan Input`.
	Cancel(i Input, err error)
}

// NewProcessor creates a process and cancel func
func NewProcessor[Input, Output any](
	process func(ctx context.Context, i Input) (Output, error),
	cancel func(i Input, err error),
) Processor[Input, Output] {
	return &processor[Input, Output]{process, cancel}
}

// processor implements Processor
type processor[Input, Output any] struct {
	process func(ctx context.Context, i Input) (Output, error)
	cancel  func(i Input, err error)
}

func (p *processor[Input, Output]) Process(ctx context.Context, i Input) (Output, error) {
	return p.process(ctx, i)
}

func (p *processor[Input, Output]) Cancel(i Input, err error) {
	p.cancel(i, err)
}
