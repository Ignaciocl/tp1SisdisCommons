package concurrency

type Semaphore chan int

// NewSemaphore initializes a semaphore with capacity equal to N
func NewSemaphore(n int) Semaphore {
	return make(Semaphore, n)
}

func (s Semaphore) Acquire() {
	s <- 1
}

func (s Semaphore) Release() {
	<-s
}
