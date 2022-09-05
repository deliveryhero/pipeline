package pipeline

import "context"

type sequence[A any] []Processor[A, A]

func (s sequence[A]) Process(ctx context.Context, a A) (A, error) {
	var zero A
	var in = a
	for _, p := range s {
		out, err := p.Process(ctx, in)
		if err != nil {
			p.Cancel(in, err)
			return zero, err
		}
		in = out
	}
	return in, nil
}

func (s sequence[A]) Cancel(_ A, _ error) {}

// Sequence connects many processors sequentially where the inputs are the same outputs
func Sequence[A any](ps ...Processor[A, A]) Processor[A, A] {
	return sequence[A](ps)
}
