package pipeline

import "context"

type apply[A, B, C any] struct {
	a Processor[A, []B]
	b Processor[B, C]
}

func (j *apply[A, B, C]) Process(ctx context.Context, a A) ([]C, error) {
	bs, err := j.a.Process(ctx, a)
	if err != nil {
		j.a.Cancel(a, err)
		return []C{}, err
	}

	cs := make([]C, 0, len(bs))

	for i := range bs {
		c, err := j.b.Process(ctx, bs[i])
		if err != nil {
			j.b.Cancel(bs[i], err)
			return cs, err
		}

		cs = append(cs, c)
	}

	return cs, nil
}

func (j *apply[A, B, C]) Cancel(_ A, _ error) {}

// Apply connects two processes, applying the second to each item of the first output
func Apply[A, B, C any](
	a Processor[A, []B],
	b Processor[B, C],
) Processor[A, []C] {
	return &apply[A, B, C]{a, b}
}
