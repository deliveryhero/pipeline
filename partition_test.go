package pipeline

import (
	"reflect"
	"testing"
)

func TestPartition(t *testing.T) {
	type args struct {
		partitions int
		route      func(i interface{}) int
		in         <-chan interface{}
	}
	tests := []struct {
		name string
		args args
		want []<-chan interface{}
	}{
		// TODO: write tests
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Partition(tt.args.partitions, tt.args.route, tt.args.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Partition() = %v, want %v", got, tt.want)
			}
		})
	}
}
