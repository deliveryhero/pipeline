package util

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestDelay(t *testing.T) {
	const maxTestDuration = time.Second
	type args struct {
		ctxTimeout time.Duration
		duration   time.Duration
		in         []interface{}
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
		name: "out closes after duration when in closes",
		args: args{
			ctxTimeout: maxTestDuration,
			duration:   maxTestDuration - 100*time.Millisecond,
			in:         []interface{}{1},
		},
		want: want{
			out:  []interface{}{1},
			open: false,
		},
	}, {
		name: "delay is not applied when the context is canceled",
		args: args{
			ctxTimeout: 10 * time.Millisecond,
			duration:   maxTestDuration,
			in:         []interface{}{1, 2, 3, 4, 5},
		},
		want: want{
			out:  []interface{}{1, 2, 3, 4, 5},
			open: false,
		},
	}, {
		name: "out is delayed by duration",
		args: args{
			ctxTimeout: maxTestDuration,
			duration:   maxTestDuration / 4,
			in:         []interface{}{1, 2, 3, 4, 5},
		},
		want: want{
			out:  []interface{}{1, 2, 3, 4},
			open: true,
		},
	}} {
		t.Run(test.name, func(t *testing.T) {
			// Create in channel
			in := make(chan interface{})
			go func() {
				defer close(in)
				for _, i := range test.args.in {
					in <- i
				}
			}()

			// Create a context with a timeut
			ctx, cancel := context.WithTimeout(context.Background(), test.args.ctxTimeout)
			defer cancel()

			// Start reading from in
			delay := Delay(ctx, test.args.duration, in)
			timeout := time.After(maxTestDuration)
			var isOpen bool
			var outs []interface{}
		loop:
			for {
				select {
				case i, open := <-delay:
					isOpen = open
					if !open {
						break loop
					}
					outs = append(outs, i)
				case <-timeout:
					break loop
				}
			}

			// Expecting the out channel to be open or closed
			if test.want.open != isOpen {
				t.Errorf("%t != %t", test.want.open, isOpen)
			}

			// Expecting processed outputs
			if !reflect.DeepEqual(test.want.out, outs) {
				t.Errorf("%+v != %+v", test.want.out, outs)
			}
		})
	}
}
