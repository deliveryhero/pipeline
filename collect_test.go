package util

import (
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
func TestCollect(t *testing.T) {
	const maxTestDuration = time.Second
	type args struct {
		maxSize     int
		maxDuration time.Duration
		in          []interface{}
		inDelay     time.Duration
	}
	type want struct {
		out  []interface{}
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
			in:          []interface{}{1, 2, 3},
			inDelay:     (maxTestDuration / 2) - (100 * time.Millisecond),
		},
		want: want{
			out:  []interface{}{[]interface{}{1, 2}},
			open: true,
		},
	}, {
		name: "collects maxSize inputs and returns",
		args: args{
			maxSize:     2,
			maxDuration: maxTestDuration / 10 * 9,
			inDelay:     maxTestDuration / 10,
			in:          []interface{}{1, 2, 3, 4, 5},
		},
		want: want{
			out: []interface{}{
				[]interface{}{1, 2},
				[]interface{}{3, 4},
				[]interface{}{5},
			},
			open: false,
		},
	}, {
		name: "collection returns after maxDuration with < maxSize",
		args: args{
			maxSize:     10,
			maxDuration: maxTestDuration / 4,
			inDelay:     (maxTestDuration / 4) - (10 * time.Millisecond),
			in:          []interface{}{1, 2, 3, 4, 5},
		},
		want: want{
			out: []interface{}{
				[]interface{}{1},
				[]interface{}{2},
				[]interface{}{3},
				[]interface{}{4},
			},
			open: true,
		},
	}} {
		t.Run(test.name, func(t *testing.T) {
			// Create the in channel
			in := make(chan interface{})
			go func() {
				defer close(in)
				for _, i := range test.args.in {
					time.Sleep(test.args.inDelay)
					in <- i
				}
			}()

			// Collect responses
			collect := Collect(test.args.maxSize, test.args.maxDuration, in)
			timeout := time.After(maxTestDuration)
			var outs []interface{}
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
				t.Errorf("%t = %t", test.want.open, isOpen)
			}

			// Expecting outputs
			if !reflect.DeepEqual(test.want.out, outs) {
				t.Errorf("%v != %v", test.want.out, outs)
			}

		})
	}
}
