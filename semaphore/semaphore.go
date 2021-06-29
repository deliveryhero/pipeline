// package semaphore is like a sync.WaitGroup with an upper limit.
// It's useful for limiting concurrent operations.
//
// Example Usage
//
//  // startMultiplying is a pipeline step that concurrently multiplies input numbers by a factor
//  func startMultiplying(concurrency, factor int, in <-chan int) <-chan int {
//  	out := make(chan int)
//  	go func() {
//   		sem := semaphore.New(concurrency)
//   		for i := range in {
//   			// Multiply up to 'concurrency' inputs at once
//   			sem.Add(1)
//   			go func() {
//   				out <- factor * i
//   				sem.Done()
//   			}()
//   		}
//   		// Wait for all multiplications to finish before closing the output chan
//   		sem.Wait()
//   		close(out)
//  	}()
//  	return out
//  }
//
package semaphore

// Semaphore is like a sync.WaitGroup, except it has a maximum
// number of items that can be added. If that maximum is reached,
// Add will block until Done is called.
type Semaphore chan struct{}

// New returns a new Semaphore
func New(max int) Semaphore {
	// There are probably more memory efficient ways to implement
	// a semaphore using runtime primitives like runtime_SemacquireMutex
	return make(Semaphore, max)
}

// Add adds delta, which may be negative, to the semaphore buffer.
// If the buffer becomes 0, all goroutines blocked by Wait are released.
// If the buffer goes negative, Add will block until another goroutine makes it positive.
// If the buffer exceeds max, Add will block until another goroutine decrements the buffer.
func (s Semaphore) Add(delta int) {
	// Increment the semaphore
	for i := delta; i > 0; i-- {
		s <- struct{}{}
	}
	// Decrement the semaphore
	for i := delta; i < 0; i++ {
		<-s
	}
}

// Done decrements the semaphore by 1
func (s Semaphore) Done() {
	s.Add(-1)
}

// Wait blocks until the semaphore is buffer is empty
func (s Semaphore) Wait() {
	// Filling the buffered channel ensures that its empty
	s.Add(cap(s))
	// Free the buffer before closing (unsure if this matters)
	s.Add(-cap(s))
	close(s)
}
