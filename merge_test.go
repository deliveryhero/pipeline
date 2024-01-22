package pipeline

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// task will wait for a specified duration of time before returning a certain number of errors
type task struct {
	id         string
	errorCount int
	waitFor    time.Duration
}

// do performs the task
func (t task) do() <-chan error {
	out := make(chan error)
	go func() {
		defer close(out)
		time.Sleep(t.waitFor)
		for i := 0; i < t.errorCount; i++ {
			out <- fmt.Errorf("[task %s] error %d", t.id, i)
		}
	}()
	return out
}

// TestMerge makes sure that the merged chan
// 1. Closes after all of its child chans close
// 2. Receives all error messages from the error chans
// 3. Stays open if one of its child chans never closes
func TestMerge(t *testing.T) {
	t.Parallel()

	maxTestDuration := time.Second
	for _, test := range []struct {
		description    string
		finishBefore   time.Duration
		expectedErrors []error
		tasks          []task
	}{{
		description:  "Closes after all of its error chans close",
		finishBefore: time.Second,
		tasks: []task{{
			id:      "a",
			waitFor: 250 * time.Millisecond,
		}, {
			id:      "b",
			waitFor: 500 * time.Millisecond,
		}, {
			id:      "c",
			waitFor: 750 * time.Millisecond,
		}},
	}, {
		description:  "Receives all errors from all of its error chans",
		finishBefore: time.Second,
		expectedErrors: []error{
			errors.New("[task a] error 0"),
			errors.New("[task c] error 0"),
			errors.New("[task c] error 1"),
			errors.New("[task c] error 2"),
			errors.New("[task b] error 0"),
			errors.New("[task b] error 1"),
		},
		tasks: []task{{
			id:         "a",
			waitFor:    250 * time.Millisecond,
			errorCount: 1,
		}, {
			id:         "b",
			waitFor:    750 * time.Millisecond,
			errorCount: 2,
		}, {
			id:         "c",
			waitFor:    500 * time.Millisecond,
			errorCount: 3,
		}},
	}, {
		description: "Stays open if one of its chans never closes",
		expectedErrors: []error{
			errors.New("[task c] error 0"),
			errors.New("[task b] error 0"),
			errors.New("[task b] error 1"),
		},
		tasks: []task{{
			id:      "a",
			waitFor: 2 * maxTestDuration,
			// We shoud expect to 'never' receive this error, because it will emit after the maxTestDuration
			errorCount: 1,
		}, {
			id:         "b",
			waitFor:    750 * time.Millisecond,
			errorCount: 2,
		}, {
			id:         "c",
			waitFor:    500 * time.Millisecond,
			errorCount: 1,
		}},
	}, {
		description: "Single channel passes through",
		expectedErrors: []error{
			errors.New("[task a] error 0"),
			errors.New("[task a] error 1"),
			errors.New("[task a] error 2"),
		},
		tasks: []task{{
			id:      "a",
			waitFor: 0,
			// We shoud expect to 'never' receive this error, because it will emit after the maxTestDuration
			errorCount: 3,
		}},
	}, {
		description:    "Closed channel returned",
		expectedErrors: []error{},
		tasks:          []task{},
	}} {
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()

			// Start doing all of the tasks
			var errChans []<-chan error
			for _, task := range test.tasks {
				errChans = append(errChans, task.do())
			}

			// Merge all of their error channels together
			var errs []error
			merged := Merge[error](errChans...)

			// Create the timeout
			timeout := time.After(maxTestDuration)
			if test.finishBefore > 0 {
				timeout = time.After(test.finishBefore)
			}

		loop:
			for {
				select {
				case i, ok := <-merged:
					if !ok {
						// The chan has closed
						break loop
					} else if err, ok := i.(error); ok {
						errs = append(errs, err)
					} else {
						t.Errorf("'%+v' is not an error!", i)
						return
					}
				case <-timeout:
					if isExpected := test.finishBefore == 0; isExpected {
						// We're testing that open channels cause a timeout
						break loop
					}
					t.Error("timed out!")
				}
			}

			// Check that all of the expected errors match the errors we received
			lenErrs, lenExpectedErros := len(errs), len(test.expectedErrors)
			for i, expectedError := range test.expectedErrors {
				if i >= lenErrs {
					t.Errorf("expectedErrors[%d]: '%s' != <nil>", i, expectedError)
				} else if err := errs[i]; expectedError.Error() != err.Error() {
					t.Errorf("expectedErrors[%d]: '%s' != %s", i, expectedError, err)
				}
			}

			// Check that we have no additional error messages other than the ones we expected
			if hasTooManyErrors := lenErrs > lenExpectedErros; hasTooManyErrors {
				for _, err := range errs[lenExpectedErros-1:] {
					t.Errorf("'%s' is unexpected!", err)
				}
			}
		})
	}

}
