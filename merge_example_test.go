package pipeline_test

import (
	"fmt"

	"github.com/deliveryhero/pipeline/v2"
)

func ExampleMerge() {
	one := pipeline.Emit(1)
	two := pipeline.Emit(2, 2)
	three := pipeline.Emit(3, 3, 3)

	for i := range pipeline.Merge(one, two, three) {
		fmt.Printf("output: %d\n", i)
	}

	fmt.Println("done")

	// Output
	// output: 1
	// output: 3
	// output: 2
	// output: 2
	// output: 3
	// output: 3
	// done
}
