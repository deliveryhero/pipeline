package pipeline

import (
	"reflect"
	"testing"
)

func TestSplit(t *testing.T) {
	type args struct {
		in interface{}
	}
	type want struct {
		out []interface{}
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

			// Collect the output
			var outs []interface{}
			for o := range Split(in) {
				outs = append(outs, o)
			}

			// Expected out
			if !reflect.DeepEqual(test.want.out, outs) {
				t.Errorf("%+v != %+v", test.want.out, outs)
			}
		})
	}
}
