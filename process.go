package util

import "context"

// Processor processes an input and reurns an output
type Processor interface {
	// Process processes an input and reurns an output
	Process(ctx context.Context, i interface{}) (interface{}, error)

	// Cancel is called if process returns an error or if the context is canceled
	Cancel(i interface{}, err error)
}

// Process processes each input and returns a cooresponding output
func Process(ctx context.Context, p Processor, in <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		// Start processing inputs until in closes
		for i := range in {
			select {
			// When the context is canceled, Cancel all inputs
			case <-ctx.Done():
				p.Cancel(i, ctx.Err())
			// Otherwise, Process all inputs
			default:
				result, err := p.Process(ctx, i)
				if err != nil {
					p.Cancel(i, err)
					continue
				}
				out <- result
			}
		}
	}()
	return out
}
