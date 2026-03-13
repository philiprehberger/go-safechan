package safechan

import "sync"

// FanIn merges multiple input channels into a single output channel.
// The output channel is closed when all input channels have been closed.
// Values are forwarded as they arrive from any input channel.
func FanIn[T any](channels ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup
	wg.Add(len(channels))

	for _, ch := range channels {
		go func(c <-chan T) {
			defer wg.Done()
			for val := range c {
				out <- val
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// FanOut distributes values from a single input channel to n output channels
// in a round-robin fashion. All output channels are closed when the input
// channel is closed.
func FanOut[T any](ch <-chan T, n int) []<-chan T {
	outs := make([]chan T, n)
	result := make([]<-chan T, n)
	for i := 0; i < n; i++ {
		outs[i] = make(chan T)
		result[i] = outs[i]
	}

	go func() {
		defer func() {
			for _, o := range outs {
				close(o)
			}
		}()
		i := 0
		for val := range ch {
			outs[i%n] <- val
			i++
		}
	}()

	return result
}

// Broadcast sends each value from the input channel to all n output channels.
// Every output channel receives every value. All output channels are closed
// when the input channel is closed.
func Broadcast[T any](ch <-chan T, n int) []<-chan T {
	outs := make([]chan T, n)
	result := make([]<-chan T, n)
	for i := 0; i < n; i++ {
		outs[i] = make(chan T)
		result[i] = outs[i]
	}

	go func() {
		defer func() {
			for _, o := range outs {
				close(o)
			}
		}()
		for val := range ch {
			for _, o := range outs {
				o <- val
			}
		}
	}()

	return result
}
