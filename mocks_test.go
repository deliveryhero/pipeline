package pipeline

import (
	"context"
	"fmt"
	"time"
)

// mockProcess is a mock of the Processor interface
type mockProcessor[Item any] struct {
	processDuration    time.Duration
	cancelDuration     time.Duration
	processReturnsErrs bool
	processed          []Item
	canceled           []Item
	errs               []string
}

// Process waits processDuration before returning its input as its output
func (m *mockProcessor[Item]) Process(ctx context.Context, i Item) (Item, error) {
	var zero Item 
	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	case <-time.After(m.processDuration):
		break
	}
	if m.processReturnsErrs {
		return zero, fmt.Errorf("process error: %v", i)
	}
	m.processed = append(m.processed, i)
	return i, nil
}

// Cancel collects all inputs that were canceled in m.canceled
func (m *mockProcessor[Item]) Cancel(i Item, err error) {
	time.Sleep(m.cancelDuration)
	m.canceled = append(m.canceled, i)
	m.errs = append(m.errs, err.Error())
}

// containsAll returns true if a and b contain all of the same elements
// in any order or if both are empty / nil
func containsAll[Item comparable](a, b []Item) bool {
	if len(a) != len(b) {
		return false
	} else if len(a) == 0 {
		return true
	}
	aMap := make(map[Item]bool)
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
