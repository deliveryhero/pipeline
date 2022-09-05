package pipeline

import (
	"context"
	"reflect"
	"testing"
	"time"
)

// TestCollect tests the following cases of the Collect func
// 1. Closes when in closes
// 2. Remains Open if in remains open
// 3. Collects max and returns immediately
// 4. Returns everything passed in if less than max after duration
// 5. After duration with nothing in the buffer, nothing is returned, channel remains open
// 6. Flushes the buffer if the context is canceled
func TestCollect(t *testing.T) {
	const maxTestDuration = time.Second
	type args struct {
		maxSize     int
		maxDuration time.Duration
		in          []int
		inDelay     time.Duration
		ctxTimeout  time.Duration
	}
	type want struct {
		out  [][]int
		open bool
	}
	for _, test := range []struct {
		name string
		args args
		want want
	}{{
		name: "out closes when in closes",
		args: args{
			maxSize:     20,
			maxDuration: maxTestDuration,
			in:          nil,
			inDelay:     0,
			ctxTimeout:  maxTestDuration,
		},
		want: want{
			out:  nil,
			open: false,
		},
	}, {
		name: "out remains open if in remains open",
		args: args{
			maxSize:     2,
			maxDuration: maxTestDuration,
			in:          []int{1, 2, 3},
			inDelay:     (maxTestDuration / 2) - (10 * time.Millisecond),
			ctxTimeout:  maxTestDuration,
		},
		want: want{
			out:  [][]int{{1, 2}},
			open: true,
		},
	}, {
		name: "collects maxSize inputs and returns",
		args: args{
			maxSize:     2,
			maxDuration: maxTestDuration / 10 * 9,
			inDelay:     maxTestDuration / 10,
			in:          []int{1, 2, 3, 4, 5},
			ctxTimeout:  maxTestDuration / 10 * 9,
		},
		want: want{
			out: [][]int{
				{1, 2},
				{3, 4},
				{5},
			},
			open: false,
		},
	}, {
		name: "collection returns after maxDuration with < maxSize",
		args: args{
			maxSize:     10,
			maxDuration: maxTestDuration / 4,
			inDelay:     (maxTestDuration / 4) - (25 * time.Millisecond),
			in:          []int{1, 2, 3, 4, 5},
			ctxTimeout:  maxTestDuration / 4,
		},
		want: want{
			out: [][]int{
				{1},
				{2},
				{3},
				{4},
			},
			open: true,
		},
	}, {
		name: "collection flushes buffer when the context is canceled",
		args: args{
			maxSize:     10,
			maxDuration: maxTestDuration,
			inDelay:     0,
			in:          []int{1, 2, 3, 4, 5},
			ctxTimeout:  0,
		},
		want: want{
			out: [][]int{
				{1, 2, 3, 4, 5},
			},
			open: false,
		},
	}} {
		t.Run(test.name, func(t *testing.T) {
			// Create the in channel
			in := make(chan int)
			go func() {
				defer close(in)
				for _, i := range test.args.in {
					time.Sleep(test.args.inDelay)
					in <- i
				}
			}()

			// Create the context
			ctx, cancel := context.WithTimeout(context.Background(), test.args.ctxTimeout)
			defer cancel()

			// Collect responses
			collect := Collect(ctx, test.args.maxSize, test.args.maxDuration, in)
			timeout := time.After(maxTestDuration)
			var outs [][]int
			var isOpen bool
		loop:
			for {
				select {
				case out, open := <-collect:
					if !open {
						isOpen = false
						break loop
					}
					isOpen = true
					outs = append(outs, out)
				case <-timeout:
					break loop
				}
			}

			// Expecting to close or stay open
			if test.want.open != isOpen {
				t.Errorf("open = %t, want %t", isOpen, test.want.open)
			}

			// Expecting outputs
			if !reflect.DeepEqual(test.want.out, outs) {
				t.Errorf("out = %v, want %v", outs, test.want.out)
			}

		})
	}
}
