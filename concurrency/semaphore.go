package concurrency

type semaphore chan int

// newSemaphore initializes a semaphore with capacity equal to N
func newSemaphore(n int) semaphore {
	return make(semaphore, n)
}

func (s semaphore) acquire() {
	s <- 1
}

func (s semaphore) release() {
	<-s
}
