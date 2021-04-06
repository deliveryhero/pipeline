package util

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestCancel(t *testing.T) {
	type args struct {
		ctx    context.Context
		cancel func(interface{}, error)
		in     <-chan interface{}
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
			if got := Cancel(tt.args.ctx, tt.args.cancel, tt.args.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cancel() = %v, want %v", got, tt.want)
			}
		})
	}
	t.Run("in never closes", func(t *testing.T) {
		// In closes after the context
		in := make(chan interface{})
		go func() {
			defer close(in)
			i := 0
			endAt := time.Now().Add(10 * time.Millisecond)
			for now := time.Now(); now.Before(endAt); now = time.Now() {
				in <- i
				i++
			}
			t.Log("ended")
		}()

		// Create a logger for the cancel fun
		canceled := func(i interface{}, err error) {
			t.Logf("canceled: %d because %s\n", i, err)
		}

		// Context times out after 1 second
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		for o := range Cancel(ctx, canceled, in) {
			t.Logf("out: %d", o)
		}

	})
}
