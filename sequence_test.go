package pipeline

import (
	"context"
	"fmt"
	"testing"
)

func TestSequence(t *testing.T) {
	// Create a step that increments the integer by 1
	inc := NewProcessor(func(_ context.Context, i int) (int, error) {
		i++
		return i, nil
	}, nil)
	// Add 5 to every number
	inc5 := Sequence(inc, inc, inc, inc, inc)
	// Verify the sequence ads 5 to each number
	var i int
	want := []int{5, 6, 7, 8, 9}
	for o := range Process(context.Background(), inc5, Emit(0, 1, 2, 3, 4)) {
		if want[i] != o {
			t.Fatalf("[%d] = %d, want %d", i, o, want[i])
		}
		i++
	}
}

func SequenceExample() {
	// Create a step that increments the integer by 1
	inc := NewProcessor(func(_ context.Context, i int) (int, error) {
		i++
		return i, nil
	}, nil)
	// Add 5 to every input
	inputs := Emit(0, 1, 2, 3, 4)
	addFive := Process(context.Background(), Sequence(inc, inc, inc, inc, inc), inputs)
	// Print the output
	for o := range addFive {
		fmt.Println(o)
	}
	// Output
	// 5
	// 6
	// 7
	// 8
	// 9
}
