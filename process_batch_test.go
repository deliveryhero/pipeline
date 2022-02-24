package pipeline

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestProcessBatch(t *testing.T) {
	const maxTestDuration = time.Second
	type args struct {
		ctxTimeout  time.Duration
		maxSize     int
		maxDuration time.Duration
		processor   *mockProcessor
		in          <-chan interface{}
	}
	tests := []struct {
		name     string
		args     args
		wantOpen bool
	}{{
		name: "out stays open if in is open",
		args: args{
			// Cancel the pipeline context half way through the test
			ctxTimeout:  maxTestDuration / 2,
			maxDuration: maxTestDuration,
			// Process 2 elements 33% of the total test duration
			maxSize: 2,
			processor: &mockProcessor{
				processDuration: maxTestDuration / 3,
				cancelDuration:  maxTestDuration / 3,
			},
			// * 10 elements = 165% of the test duration
			in: Emit[interface{}](1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
		},
		// Therefore the out chan should still be open when the test times out
		wantOpen: true,
	}, {
		name: "out closes if in is closed",
		args: args{
			// Cancel the pipeline context half way through the test
			ctxTimeout:  maxTestDuration / 2,
			maxDuration: maxTestDuration,
			// Process 5 elements 33% of the total test duration
			maxSize: 5,
			processor: &mockProcessor{
				processDuration: maxTestDuration / 3,
				cancelDuration:  maxTestDuration / 3,
			},
			// * 10 elements = 66% of the test duration
			in: Emit[interface{}](1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
		},
		// Therefore the out channel should be closed when the test ends
		wantOpen: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancel := context.WithTimeout(context.Background(), tt.args.ctxTimeout)
			defer cancel()

			// Process the batch with a timeout of maxTestDuration
			open := true
			outChan := ProcessBatch(ctx, tt.args.maxSize, tt.args.maxDuration, tt.args.processor, tt.args.in)
			timeout := time.After(maxTestDuration)
		loop:
			for {
				select {
				case <-timeout:
					break loop
				case _, ok := <-outChan:
					if !ok {
						open = false
						break loop
					}
				}
			}
			// Expecting the channels open state
			if open != tt.wantOpen {
				t.Errorf("open = %t, wanted %t", open, tt.wantOpen)
			}
		})
	}
}

func TestProcessBatchConcurrently(t *testing.T) {
	const maxTestDuration = time.Second
	type args struct {
		ctxTimeout   time.Duration
		concurrently int
		maxSize      int
		maxDuration  time.Duration
		processor    *mockProcessor
		in           <-chan interface{}
	}
	tests := []struct {
		name     string
		args     args
		wantOpen bool
	}{{
		name: "out stays open if in is open",
		args: args{
			ctxTimeout:  maxTestDuration / 2,
			maxDuration: maxTestDuration,
			// Process 1 element for 33% of the total test duration
			maxSize: 1,
			// * 2x concurrently
			concurrently: 2,
			processor: &mockProcessor{
				processDuration: maxTestDuration / 3,
				cancelDuration:  maxTestDuration / 3,
			},
			// * 10 elements = 165% of the test duration
			in: Emit[interface{}](1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
		},
		// Therefore the out chan should still be open when the test times out
		wantOpen: true,
	}, {
		name: "out closes if in is closed",
		args: args{
			ctxTimeout:  maxTestDuration / 2,
			maxDuration: maxTestDuration,
			// Process 1 element for 33% of the total test duration
			maxSize: 1,
			// * 5x concurrently
			concurrently: 5,
			processor: &mockProcessor{
				processDuration: maxTestDuration / 3,
				cancelDuration:  maxTestDuration / 3,
			},
			// * 10 elements = 66% of the test duration
			in: Emit[interface{}](1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
		},
		// Therefore the out channel should be closed by the end of the test
		wantOpen: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancel := context.WithTimeout(context.Background(), tt.args.ctxTimeout)
			defer cancel()

			// Process the batch with a timeout of maxTestDuration
			open := true
			out := ProcessBatchConcurrently(ctx, tt.args.concurrently, tt.args.maxSize, tt.args.maxDuration, tt.args.processor, tt.args.in)
			timeout := time.After(maxTestDuration)
		loop:
			for {
				select {
				case <-timeout:
					break loop
				case _, ok := <-out:
					if !ok {
						open = false
						break loop
					}
				}
			}
			// Expecting the channels open state
			if open != tt.wantOpen {
				t.Errorf("open = %t, wanted %t", open, tt.wantOpen)
			}
		})
	}
}

func Test_processBatch(t *testing.T) {
	drain := make(chan interface{}, 10000)
	const maxTestDuration = time.Second
	type args struct {
		ctxTimeout  time.Duration
		maxSize     int
		maxDuration time.Duration
		processor   *mockProcessor
		in          <-chan interface{}
		out         chan<- interface{}
	}
	type want struct {
		open      bool
		processed []interface{}
		canceled  []interface{}
		errs      []interface{}
	}
	tests := []struct {
		name string
		args args
		want want
	}{{
		name: "returns instantly if in is closed",
		args: args{
			ctxTimeout:  maxTestDuration,
			maxSize:     20,
			maxDuration: maxTestDuration,
			processor:   new(mockProcessor),
			in: func() <-chan interface{} {
				in := make(chan interface{})
				close(in)
				return in
			}(),
			out: drain,
		},
		want: want{
			open: false,
		},
	}, {
		name: "processes slices of inputs",
		args: args{
			ctxTimeout:  maxTestDuration,
			maxSize:     2,
			maxDuration: maxTestDuration,
			processor:   new(mockProcessor),
			in:          Emit[interface{}](1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
			out:         drain,
		},
		want: want{
			open: false,
			processed: []interface{}{[]interface{}{
				1, 2,
			}, []interface{}{
				3, 4,
			}, []interface{}{
				5, 6,
			}, []interface{}{
				7, 8,
			}, []interface{}{
				9, 10,
			}},
		},
	}, {
		name: "cancels slices of inputs if process returns an error",
		args: args{
			ctxTimeout:  maxTestDuration / 2,
			maxSize:     5,
			maxDuration: maxTestDuration,
			processor: &mockProcessor{
				processReturnsErrs: true,
			},
			in:  Emit[interface{}](1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
			out: drain,
		},
		want: want{
			open: false,
			canceled: []interface{}{[]interface{}{
				1, 2, 3, 4, 5,
			}, []interface{}{
				6, 7, 8, 9, 10,
			}},
			errs: []interface{}{
				"process error: [1 2 3 4 5]",
				"process error: [6 7 8 9 10]",
			},
		},
	}, {
		name: "cancels slices of inputs when the context is canceled",
		args: args{
			ctxTimeout:  maxTestDuration / 2,
			maxSize:     1,
			maxDuration: maxTestDuration,
			processor: &mockProcessor{
				// this will take longer to complete than the maxTestDuration by a few micro seconds
				processDuration: maxTestDuration / 10,                     // 5 calls to Process > maxTestDuration / 2
				cancelDuration:  maxTestDuration/10 + 25*time.Millisecond, // 5 calls to Cancel >  maxTestDuration / 2
			},
			in:  Emit[interface{}](1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
			out: drain,
		},
		want: want{
			open: true,
			processed: []interface{}{
				[]interface{}{1},
				[]interface{}{2},
				[]interface{}{3},
				[]interface{}{4},
			},
			canceled: []interface{}{
				[]interface{}{5},
				[]interface{}{6},
				[]interface{}{7},
				[]interface{}{8},
			},
			errs: []interface{}{
				"context deadline exceeded",
				"context deadline exceeded",
				"context deadline exceeded",
				"context deadline exceeded",
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.args.ctxTimeout)
			defer cancel()

			// Process the batch with a timeout of maxTestDuration
			timeout := time.After(maxTestDuration)
			open := true
		loop:
			for {
				select {
				case <-timeout:
					break loop
				default:
					open = processOneBatch(ctx, tt.args.maxSize, tt.args.maxDuration, tt.args.processor, tt.args.in, tt.args.out)
					if !open {
						break loop
					}
				}
			}

			// Processing took longer than expected
			if open != tt.want.open {
				t.Errorf("open = %t, wanted %t", open, tt.want.open)
			}
			// Expecting processed inputs
			if !reflect.DeepEqual(tt.args.processor.processed, tt.want.processed) {
				t.Errorf("processed = %+v, want %+v", tt.args.processor.processed, tt.want.processed)
			}
			// Expecting canceled inputs
			if !reflect.DeepEqual(tt.args.processor.canceled, tt.want.canceled) {
				t.Errorf("canceled = %+v, want %+v", tt.args.processor.canceled, tt.want.canceled)
			}
			// Expecting canceled errors
			if !reflect.DeepEqual(tt.args.processor.errs, tt.want.errs) {
				t.Errorf("errs = %+v, want %+v", tt.args.processor.errs, tt.want.errs)
			}
		})
	}
}
