package pipeline

import (
	"context"
	"fmt"
	"time"
)

// mockProcess is a mock of the Processor interface
type mockProcessor struct {
	processDuration    time.Duration
	cancelDuration     time.Duration
	processReturnsErrs bool
	processed          []interface{}
	canceled           []interface{}
	errs               []interface{}
}

// Process waits processDuration before returning its input at its output
func (m *mockProcessor) Process(ctx context.Context, i interface{}) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(m.processDuration):
		break
	}
	if m.processReturnsErrs {
		return nil, fmt.Errorf("process error: %d", i)
	}
	m.processed = append(m.processed, i)
	return i, nil
}

// Cancel collects all inputs that were canceled in m.canceled
func (m *mockProcessor) Cancel(i interface{}, err error) {
	time.Sleep(m.cancelDuration)
	m.canceled = append(m.canceled, i)
	m.errs = append(m.errs, err.Error())
}

// containsAll returns true if a and b contain all of the same elements
// in any order or if both are empty / nil
func containsAll(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	} else if len(a) == 0 {
		return true
	}
	aMap := make(map[interface{}]bool)
	for _, i := range a {
		aMap[i] = true
	}
	for _, i := range b {
		if !aMap[i] {
			return false
		}
	}
	return true
}
