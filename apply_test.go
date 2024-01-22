package pipeline

import (
	"context"
	"strings"
	"testing"
)

func TestLoopApply(t *testing.T) {
	transform := NewProcessor(func(_ context.Context, s string) ([]string, error) {
		return strings.Split(s, ","), nil
	}, nil)

	double := NewProcessor(func(_ context.Context, s string) (string, error) {
		return s + s, nil
	}, nil)

	addLeadingZero := NewProcessor(func(_ context.Context, s string) (string, error) {
		return "0" + s, nil
	}, nil)

	looper := Apply(
		transform,
		Sequence(
			double,
			addLeadingZero,
			double,
		),
	)

	gotCount := 0
	input := "1,2,3,4,5"
	want := []string{"011011", "022022", "033033", "044044", "055055"}

	for out := range Process(context.Background(), looper, Emit(input)) {
		for j := range out {
			gotCount++
			if !contains(want, out[j]) {
				t.Errorf("does not contains got=%v, want=%v", out[j], want)
			}
		}
	}

	if gotCount != len(want) {
		t.Errorf("total results got=%v, want=%v", gotCount, len(want))
	}
}

func contains(s []string, e string) bool {
	for i := range s {
		if s[i] == e {
			return true
		}
	}

	return false
}
