package util

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestProcessBatch(t *testing.T) {
	type args struct {
		ctx         context.Context
		maxSize     int
		maxDuration time.Duration
		processor   Processor
		in          <-chan interface{}
	}
	tests := []struct {
		name string
		args args
		want <-chan interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProcessBatch(tt.args.ctx, tt.args.maxSize, tt.args.maxDuration, tt.args.processor, tt.args.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessBatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
