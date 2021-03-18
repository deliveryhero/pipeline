package util

import (
	"reflect"
	"testing"
)

func TestSplit(t *testing.T) {
	type args struct {
		in interface{}
	}
	type want struct {
		out   []interface{}
		panic bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"splits slices if interfaces into individual interfaces",
			args{
				in: []interface{}{1, 2, 3, 4, 5},
			},
			want{
				out: []interface{}{1, 2, 3, 4, 5},
			},
		}, {
			"split panics if it isn't given a slice of interfaces",
			args{
				in: []int{1, 2, 3, 4, 5},
			},
			want{
				panic: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create the in channel
			in := make(chan interface{})
			go func() {
				defer close(in)
				in <- test.args.in
			}()

			// Recover from panics
			defer func() {
				// TODO: find a working approach to capturing a panic in this case
				if err := recover(); err != nil {
					if !test.want.panic {
						t.Errorf("panic: %s", err)
					}
				}
			}()

			// Collect the output
			var outs []interface{}
			for o := range Split(in) {
				outs = append(outs, o)
			}

			// Expected to panic
			if test.want.panic {
				t.Error("expected to panic")
			}

			// Expected out
			if !reflect.DeepEqual(test.want.out, outs) {
				t.Errorf("%+v != %+v", test.want.out, outs)
			}
		})
	}
}
