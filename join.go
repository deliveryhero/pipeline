package pipeline

import "context"

type join[A, B, C any] struct {
	a Processor[A, B]
	b Processor[B, C]
}

func (j *join[A, B, C]) Process(ctx context.Context, a A) (C, error) {
	var zero C
	if b, err := j.a.Process(ctx, a); err != nil {
		j.a.Cancel(a, err)
		return zero, err
	} else if c, err := j.b.Process(ctx, b); err != nil {
		j.b.Cancel(b, err)
		return zero, err
	} else {
		return c, nil
	}
}

func (j *join[A, B, C]) Cancel(_ A, _ error) {}

// Join connects two processes where the output of the first is the input of the second
func Join[A, B, C any](a Processor[A, B], b Processor[B, C]) Processor[A, C] {
	return &join[A, B, C]{a, b}
}
