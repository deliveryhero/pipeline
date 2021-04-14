// db is a fake db package for use with some of the examples
package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Result is returned from the 'database'
type Result struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

func GetAdvertisements(ctx context.Context, query string, rs *[]Result) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		if err := simulatedDbRequest(ctx, "advertisement", rs); err != nil {
			out <- err
		}
	}()
	return out
}

func GetImages(ctx context.Context, query string, rs *[]Result) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		if err := simulatedDbRequest(ctx, "image", rs); err != nil {
			out <- err
		}
	}()
	return out
}

func GetProducts(ctx context.Context, query string, rs *[]Result) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		if err := simulatedDbRequest(ctx, "product", rs); err != nil {
			out <- err
		}
	}()
	return out
}

func GetWebsites(ctx context.Context, query string, rs *[]Result) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		if err := simulatedDbRequest(ctx, "website", rs); err != nil {
			out <- err
		}
	}()
	return out
}

func simulatedDbRequest(ctx context.Context, resultPrefix string, rs *[]Result) error {
	select {
	// Return right away if the
	case <-ctx.Done():
		return ctx.Err()
	// Simulate 25-200ms of latency from a db request
	case <-time.After(time.Millisecond * time.Duration((25 + (rand.Int() % 175)))):
		break
	}
	// Return between 1 and 5 results
	total := 1 + (rand.Int() % 5)
	for i := 0; i < total; i++ {
		*rs = append(*rs, Result{
			Title:       fmt.Sprintf("%s %d title", resultPrefix, i),
			Description: fmt.Sprintf("%s %d description", resultPrefix, i),
			URL:         fmt.Sprintf("http://%s-%d.com", resultPrefix, i),
		})
	}
	return nil
}
