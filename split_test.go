package pipeline

import (
	"reflect"
	"testing"
)

func TestSplit(t *testing.T) {
	type args struct {
		in []int
	}
	type want struct {
		out []int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"splits slices if interfaces into individual interfaces",
			args{
				in: []int{1, 2, 3, 4, 5},
			},
			want{
				out: []int{1, 2, 3, 4, 5},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create the in channel
			in := Emit(test.args.in)

			// Collect the output
			var outs []int
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
