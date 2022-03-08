package pipeline_test

import (
	"fmt"

	pipeline "github.com/deliveryhero/pipeline/v2-beta"
)

func ExampleMerge() {
	one := pipeline.Emit[int](1)
	two := pipeline.Emit[int](2, 2)
	three := pipeline.Emit[int](3, 3, 3)

	for i := range pipeline.Merge[int](one, two, three) {
		fmt.Printf("output: %d\n", i)
	}

	fmt.Println("done")

	// Output:
	// output: 3
	// output: 3
	// output: 2
	// output: 1
	// output: 2
	// output: 3
	// done
}
