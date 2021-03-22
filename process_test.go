package util

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
)

// mockProcess is a mock of the Processor interface
type mockProcessor struct {
	processDuration    time.Duration
	processReturnsErrs bool
	cancelDuration     time.Duration
	canceled           []interface{}
	canceledErrs       []string
}

// Process waits processDuration before returning its input at its output
func (m *mockProcessor) Process(i interface{}) (interface{}, error) {
	time.Sleep(m.processDuration)
	if m.processReturnsErrs {
		return nil, fmt.Errorf("process error: %d", i)
	}
	return i, nil
}

// Cancel collects all inputs that were canceled in m.canceled
func (m *mockProcessor) Cancel(i interface{}, err error) {
	time.Sleep(m.cancelDuration)
	m.canceled = append(m.canceled, i)
	m.canceledErrs = append(m.canceledErrs, err.Error())
}

func TestProcess(t *testing.T) {
	const maxTestDuration = time.Second
	type args struct {
		ctxTimeout           time.Duration
		processDuration      time.Duration
		processReturnsErrors bool
		cancelDuration       time.Duration
		in                   []interface{}
	}
	type want struct {
		open         bool
		out          []interface{}
		canceled     []interface{}
		canceledErrs []string
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
				canceledErrs: []string{
					"context deadline exceeded",
					"context deadline exceeded",
					"context deadline exceeded",
					"context deadline exceeded",
					"context deadline exceeded",
				},
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
				canceledErrs: []string{
					"context deadline exceeded",
				},
			},
		}, {
			name: "when an error is returned during process, it is passed to cancel",
			args: args{
				ctxTimeout:           maxTestDuration - 100*time.Millisecond,
				processDuration:      (maxTestDuration - 200*time.Millisecond) / 2,
				processReturnsErrors: true,
				cancelDuration:       0,
				in:                   []interface{}{1, 2, 3},
			},
			want: want{
				open:     false,
				out:      nil,
				canceled: []interface{}{1, 2, 3},
				canceledErrs: []string{
					"process error: 1",
					"process error: 2",
					"context deadline exceeded",
				},
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
				processDuration:    test.args.processDuration,
				processReturnsErrs: test.args.processReturnsErrors,
				cancelDuration:     test.args.cancelDuration,
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

			// Expecting canceled errors
			if !reflect.DeepEqual(test.want.canceledErrs, processor.canceledErrs) {
				t.Errorf("%+v != %+v", test.want.canceledErrs, processor.canceledErrs)
			}
		})
	}
}
