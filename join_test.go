package pipeline

import (
	"context"
	"fmt"
	"strconv"
	"testing"
)

func TestJoin(t *testing.T) {
	// Emit 10 numbers
	want := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	ins := Emit(want...)
	// Join two steps, one that converts the number to a string, the other that converts it back to a number
	join := Join(NewProcessor(func(_ context.Context, i int) (string, error) {
		return strconv.Itoa(i), nil
	}, nil), NewProcessor(func(_ context.Context, i string) (int, error) {
		return strconv.Atoi(i)
	}, nil))
	// Compare inputs and outputs
	var idx int
	for got := range Process(context.Background(), join, ins) {
		if want[idx] != got {
			t.Fatalf("[%d] = %d, want = %d", idx, got, want[idx])
		}
		idx++
	}
}

func JoinExample() {
	// Emit 10 numbers
	inputs := Emit(0, 1, 2, 3, 4, 5)
	// Join two steps, one that converts the number to a string, the other that converts it back to a number
	convertToStringThenBackToInt := Process(context.Background(), Join(NewProcessor(func(_ context.Context, i int) (string, error) {
		return strconv.Itoa(i), nil
	}, nil), NewProcessor(func(_ context.Context, i string) (int, error) {
		return strconv.Atoi(i)
	}, nil)), inputs)
	// Print the output
	for o := range convertToStringThenBackToInt {
		fmt.Println(o)
	}
	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
	// 5
}
