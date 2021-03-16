package util

import (
	"reflect"
	"testing"
	"time"
)

// TestCollect tests the following cases of the Collect func
// 1. Closes when in closes
// 2. Remains Open if in remains open
// 3. Collects max and returns chunks
// 4. Returns everything passed in if less than max after duration
// 5. After duration with nothing in the buffer, nothing is returned, channel remains open
func TestCollect(t *testing.T) {
	// emit emits a slice of strings as a channel of interfaces
	emit := func(ins []string, delay time.Duration) <-chan interface{} {
		out := make(chan interface{})
		go func() {
			defer close(out)
			for _, i := range ins {
				time.Sleep(delay)
				out <- i
			}
		}()
		return out
	}

	type args struct {
		max      int
		duration time.Duration
		in       []string
		inDelay  time.Duration
	}
	type want struct {
		out   [][]interface{}
		close bool
	}
	for _, test := range []struct {
		name    string
		timeout time.Duration
		args    args
		want    want
	}{{
		name:    "closes when in closes",
		timeout: 10 * time.Millisecond,
		args: args{
			max:      20,
			duration: time.Second, // should close even if duration hasn't elapsed
			in:       nil,
		},
		want: want{
			out:   nil,
			close: true,
		},
	}, {
		name:    "remains open if in remains open",
		timeout: time.Second,
		args: args{
			max:      1,
			duration: time.Second,
			in:       []string{"one second", "two seconds", "three seconds"},
			inDelay:  750 * time.Millisecond,
		},
		want: want{
			out:   [][]interface{}{{"one second"}},
			close: false,
		},
	}, {
		name:    "collects max and returns chunks",
		timeout: 60 * time.Millisecond,
		args: args{
			max:      2,
			duration: time.Second,
			inDelay:  10 * time.Millisecond,
			in:       []string{"1", "2", "3", "4", "5"},
		},
		want: want{
			out:   [][]interface{}{{"1", "2"}, {"3", "4"}, {"5"}},
			close: true,
		},
	}, {
		name:    "returns chunk after duration",
		timeout: 45 * time.Millisecond,
		args: args{
			max:      10,
			duration: 10 * time.Millisecond,
			inDelay:  8 * time.Millisecond,
			in:       []string{"1", "2", "3", "4", "5"},
		},
		want: want{
			out:   [][]interface{}{{"1"}, {"2"}, {"3"}, {"4"}},
			close: false,
		},
	}} {
		t.Run(test.name, func(t *testing.T) {
			// Create the timeout
			timeoutAfter := time.After(test.timeout)

			// Collect responses
			collect := Collect(test.args.max, test.args.duration, emit(test.args.in, test.args.inDelay))
			var outs [][]interface{}
			var didClose bool
		loop:
			for {
				select {
				case out, open := <-collect:
					if !open {
						didClose = true
						break loop
					}
					outs = append(outs, out)
				case <-timeoutAfter:
					break loop
				}
			}

			// Did we close? Were we expecting to close
			if test.want.close && !didClose {
				t.Errorf("expected out channel to close!")
			} else if !test.want.close && didClose {
				t.Errorf("did not expect out channel to close!")
			}

			// Compare outputs
			if !reflect.DeepEqual(test.want.out, outs) {
				t.Errorf("%v != %v", test.want.out, outs)
			}

		})
	}
}
