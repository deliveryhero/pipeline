package pipeline_test

import (
	"fmt"
	"testing"

	"github.com/deliveryhero/pipeline/v2"
)

func TestExampleMerge(t *testing.T) {
	one := pipeline.Emit(1)
	two := pipeline.Emit(2, 2)
	three := pipeline.Emit(3, 3, 3)

	for i := range pipeline.Merge(one, two, three) {
		fmt.Printf("output: %d\n", i)
	}

	fmt.Println("done")

	// Output:
	// Output:: 1
	// Output:: 3
	// Output:: 2
	// Output:: 2
	// Output:: 3
	// Output:: 3
	// done
}
