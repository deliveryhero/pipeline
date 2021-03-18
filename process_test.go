package util

import (
	"context"
	"reflect"
	"testing"
	"time"
)

// mockProcess is a mock of the Processor interface
type mockProcessor struct {
	processDuration time.Duration
	cancelDuration  time.Duration
	canceled        []interface{}
}

// Process waits processDuration before returning its input at its output
func (m *mockProcessor) Process(i interface{}) interface{} {
	time.Sleep(m.processDuration)
	return i
}

// Cancel collects all inputs that were canceled in m.canceled
func (m *mockProcessor) Cancel(i interface{}) {
	time.Sleep(m.cancelDuration)
	m.canceled = append(m.canceled, i)
}

func TestProcess(t *testing.T) {
	const maxTestDuration = time.Second
	type args struct {
		ctxTimeout      time.Duration
		processDuration time.Duration
		cancelDuration  time.Duration
		in              []interface{}
	}
	type want struct {
		open     bool
		out      []interface{}
		canceled []interface{}
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "out closes if in closes but the context isn't canceled",
			args: args{
				ctxTimeout:      2 * maxTestDuration,
				processDuration: 0,
				in:              []interface{}{1, 2, 3},
			},
			want: want{
				open:     false,
				out:      []interface{}{1, 2, 3},
				canceled: nil,
			},
		}, {
			name: "cancel is called on elements after the context is canceled",
			args: args{
				ctxTimeout:      maxTestDuration / 2,
				processDuration: maxTestDuration / 11,
				in:              []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
			want: want{
				open:     false,
				out:      []interface{}{1, 2, 3, 4, 5},
				canceled: []interface{}{6, 7, 8, 9, 10},
			},
		}, {
			name: "out stays open as long as in is open",
			args: args{
				ctxTimeout:      maxTestDuration / 2,
				processDuration: (maxTestDuration / 2) - (100 * time.Millisecond),
				cancelDuration:  (maxTestDuration / 2) - (100 * time.Millisecond),
				in:              []interface{}{1, 2, 3},
			},
			want: want{
				open:     true,
				out:      []interface{}{1},
				canceled: []interface{}{2},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create the in channel
			in := make(chan interface{})
			go func() {
				defer close(in)
				for _, i := range test.args.in {
					in <- i
				}
			}()

			// Setup the Processor
			ctx, cancel := context.WithTimeout(context.Background(), test.args.ctxTimeout)
			defer cancel()
			processor := &mockProcessor{
				processDuration: test.args.processDuration,
				cancelDuration:  test.args.cancelDuration,
			}
			out := Process(ctx, processor, in)

			// Collect the outputs
			timeout := time.After(maxTestDuration)
			var outs []interface{}
			var isOpen bool
		loop:
			for {
				select {
				case o, open := <-out:
					if !open {
						isOpen = false
						break loop
					}
					isOpen = true
					outs = append(outs, o)
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

			// Expecting canceled inputs
			if !reflect.DeepEqual(test.want.canceled, processor.canceled) {
				t.Errorf("%+v != %+v", test.want.canceled, processor.canceled)
			}
		})
	}
}
