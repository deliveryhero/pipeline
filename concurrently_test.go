package util_test

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/marksalpeter/util"
)

func TestConcurrently(t *testing.T) {
	timeoutDurration := time.Second
	for _, test := range []struct {
		description   string
		inputs        []string
		expects       []string
		intputDelay   time.Duration
		concurrently  int
		shouldTimeout bool
	}{{
		description:  "outs all close if the in closes",
		concurrently: 2,
		intputDelay:  100 * time.Millisecond,
	}, {
		description:  "each out receives inputs from in",
		inputs:       []string{"one", "two", "three", "four", "five"},
		expects:      []string{"one", "two", "three", "four", "five"},
		concurrently: 2,
		intputDelay:  100 * time.Millisecond,
	}, {
		description:   "if the in blocks, the outs never close",
		concurrently:  3,
		inputs:        []string{"one", "two"},
		shouldTimeout: true,
		intputDelay:   10 * timeoutDurration,
	}} {
		t.Run(test.description, func(t *testing.T) {
			// Setup the input channel
			in := make(chan interface{})
			go func() {
				defer close(in)
				for _, input := range test.inputs {
					// Preserve the order order
					time.Sleep(test.intputDelay)
					in <- input
				}
			}()

			// Fan out the input channel using concurrently
			outs := util.Concurrently(test.concurrently, in)

			// Incorrect number of output channels
			if l := len(outs); l != test.concurrently {
				t.Errorf("%d != %d", l, test.concurrently)
				return
			}

			// Consume each output chan
			var mutex sync.Mutex
			var outputs []string
			didChanReceiveOutput := make([]bool, len(outs))
			didChanClose := make([]bool, len(outs))
			for i := range outs {
				go func(i int, out <-chan interface{}) {
					for o := range out {
						if o == nil {
							continue
						}
						// The channel received an message
						mutex.Lock()
						outputs = append(outputs, o.(string))
						didChanReceiveOutput[i] = true
						mutex.Unlock()
					}

					// The channel closed
					mutex.Lock()
					didChanClose[i] = true
					mutex.Unlock()
				}(i, outs[i])
			}

			// Wait until we're done consuming the outputs
			time.Sleep(timeoutDurration)

			// Did every out chan receive something if we expected it to?
			for outputChanIndex, didReceiveOutput := range didChanReceiveOutput {
				shouldHaveReceivedInput := outputChanIndex < len(test.expects)
				if shouldHaveReceivedInput && !didReceiveOutput {
					t.Errorf("[chan %d] did not receive output", outputChanIndex)
				}
			}

			// Did every chan close that we expected to?
			// Did evey chan remain open that we expected to?
			for outputChanIndex, didChanClose := range didChanClose {
				if test.shouldTimeout && didChanClose {
					t.Errorf("[chan %d] did not timeout", outputChanIndex)
				} else if !test.shouldTimeout && !didChanClose {
					t.Errorf("[chan %d] did not close", outputChanIndex)
				}
			}

			// Do the inputs match the expected outputs?
			if !reflect.DeepEqual(test.expects, outputs) {
				t.Errorf("%+v != %+v", test.expects, outputs)
			}

		})
	}
}
