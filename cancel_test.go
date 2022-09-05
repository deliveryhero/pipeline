package pipeline

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCancel(t *testing.T) {
	const testDuration = time.Second

	// Collect logs
	var logs []string
	logf := func(v string, is ...interface{}) {
		logs = append(logs, fmt.Sprintf(v, is...))
	}

	// Send a stream of ints through the in chan
	in := make(chan int)
	go func() {
		defer close(in)
		i := 0
		endAt := time.Now().Add(testDuration)
		for now := time.Now(); now.Before(endAt); now = time.Now() {
			in <- i
			i++
			time.Sleep(testDuration / 100)
		}
		logf("ended")
	}()

	// Create a logger for the cancel func
	canceled := func(i int, err error) {
		logf("canceled: %d because %s\n", i, err)
	}

	// Start canceling the pipeline about half way through the test
	ctx, cancel := context.WithTimeout(context.Background(), testDuration/2)
	defer cancel()
	for i := range Cancel(ctx, canceled, in) {
		logf("%d", i)
	}

	// There should be some logs
	lenLogs := len(logs)
	if lenLogs < 2 {
		t.Errorf("len(logs) = %d, wanted > 2", lenLogs)
		t.Log(logs)
		return
	}

	// The first half of the logs (+-20%) should be a string representation of the numbers in order
	var iCanceled int
	for i, log := range logs {
		if strconv.Itoa(i) != log {
			iCanceled = i
			if isAboutHalfWay := math.Abs(float64((lenLogs/2)-i)) <= .2*float64(lenLogs); isAboutHalfWay {
				break
			}
			t.Errorf("got %d, wanted %s", i, log)
			for _, l := range logs {
				t.Error(l)
			}
		}
	}

	// The remaining logs should be prefixed with "canceled:"
	for i, log := range logs[iCanceled : lenLogs-1] {
		if !strings.Contains(log, "canceled:") {
			t.Errorf("got '%s', wanted 'canceled: %d because context deadline exceeded'", log, i+lenLogs/2)
		}
	}

}
