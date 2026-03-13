package safechan

import (
	"sort"
	"sync"
	"testing"
	"time"
)

func TestFanIn_BasicMerge(t *testing.T) {
	ch1 := make(chan int, 2)
	ch2 := make(chan int, 2)

	ch1 <- 1
	ch1 <- 2
	close(ch1)

	ch2 <- 3
	ch2 <- 4
	close(ch2)

	out := FanIn(ch1, ch2)
	var results []int
	for val := range out {
		results = append(results, val)
	}

	sort.Ints(results)
	if len(results) != 4 {
		t.Fatalf("expected 4 values, got %d", len(results))
	}
	expected := []int{1, 2, 3, 4}
	for i, v := range expected {
		if results[i] != v {
			t.Fatalf("expected %d at index %d, got %d", v, i, results[i])
		}
	}
}

func TestFanIn_SingleChannel(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 10
	ch <- 20
	ch <- 30
	close(ch)

	out := FanIn(ch)
	var results []int
	for val := range out {
		results = append(results, val)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 values, got %d", len(results))
	}
}

func TestFanIn_NoChannels(t *testing.T) {
	out := FanIn[int]()
	_, ok := <-out
	if ok {
		t.Fatal("expected output channel to be closed immediately with no inputs")
	}
}

func TestFanIn_ClosesOutput(t *testing.T) {
	ch := make(chan string)
	close(ch)
	out := FanIn(ch)
	_, ok := <-out
	if ok {
		t.Fatal("expected output channel to close when input closes")
	}
}

func TestFanIn_ConcurrentProducers(t *testing.T) {
	const numChannels = 5
	const msgsPerChannel = 20

	channels := make([]<-chan int, numChannels)
	for i := 0; i < numChannels; i++ {
		ch := make(chan int)
		channels[i] = ch
		go func(c chan int, start int) {
			for j := 0; j < msgsPerChannel; j++ {
				c <- start + j
			}
			close(c)
		}(ch, i*msgsPerChannel)
	}

	out := FanIn(channels...)
	var results []int
	for val := range out {
		results = append(results, val)
	}

	if len(results) != numChannels*msgsPerChannel {
		t.Fatalf("expected %d values, got %d", numChannels*msgsPerChannel, len(results))
	}
}

func TestFanOut_RoundRobin(t *testing.T) {
	in := make(chan int, 6)
	for i := 0; i < 6; i++ {
		in <- i
	}
	close(in)

	outs := FanOut(in, 3)
	if len(outs) != 3 {
		t.Fatalf("expected 3 output channels, got %d", len(outs))
	}

	// Collect results from each output channel
	results := make([][]int, 3)
	var wg sync.WaitGroup
	for i, ch := range outs {
		wg.Add(1)
		go func(idx int, c <-chan int) {
			defer wg.Done()
			for val := range c {
				results[idx] = append(results[idx], val)
			}
		}(i, ch)
	}
	wg.Wait()

	// Channel 0 should get 0, 3; Channel 1 should get 1, 4; Channel 2 should get 2, 5
	expected := [][]int{{0, 3}, {1, 4}, {2, 5}}
	for i, exp := range expected {
		if len(results[i]) != len(exp) {
			t.Fatalf("channel %d: expected %d values, got %d", i, len(exp), len(results[i]))
		}
		for j, v := range exp {
			if results[i][j] != v {
				t.Fatalf("channel %d index %d: expected %d, got %d", i, j, v, results[i][j])
			}
		}
	}
}

func TestFanOut_SingleOutput(t *testing.T) {
	in := make(chan int, 3)
	in <- 1
	in <- 2
	in <- 3
	close(in)

	outs := FanOut(in, 1)
	if len(outs) != 1 {
		t.Fatalf("expected 1 output channel, got %d", len(outs))
	}

	var results []int
	for val := range outs[0] {
		results = append(results, val)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 values, got %d", len(results))
	}
}

func TestFanOut_ClosesOutputs(t *testing.T) {
	in := make(chan int)
	close(in)

	outs := FanOut(in, 2)
	for i, ch := range outs {
		_, ok := <-ch
		if ok {
			t.Fatalf("expected output channel %d to be closed", i)
		}
	}
}

func TestBroadcast_AllReceive(t *testing.T) {
	in := make(chan int, 3)
	in <- 10
	in <- 20
	in <- 30
	close(in)

	outs := Broadcast(in, 3)
	if len(outs) != 3 {
		t.Fatalf("expected 3 output channels, got %d", len(outs))
	}

	results := make([][]int, 3)
	var wg sync.WaitGroup
	for i, ch := range outs {
		wg.Add(1)
		go func(idx int, c <-chan int) {
			defer wg.Done()
			for val := range c {
				results[idx] = append(results[idx], val)
			}
		}(i, ch)
	}
	wg.Wait()

	expected := []int{10, 20, 30}
	for i := 0; i < 3; i++ {
		if len(results[i]) != 3 {
			t.Fatalf("output %d: expected 3 values, got %d", i, len(results[i]))
		}
		for j, v := range expected {
			if results[i][j] != v {
				t.Fatalf("output %d index %d: expected %d, got %d", i, j, v, results[i][j])
			}
		}
	}
}

func TestBroadcast_SingleOutput(t *testing.T) {
	in := make(chan string, 2)
	in <- "a"
	in <- "b"
	close(in)

	outs := Broadcast(in, 1)
	var results []string
	for val := range outs[0] {
		results = append(results, val)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 values, got %d", len(results))
	}
	if results[0] != "a" || results[1] != "b" {
		t.Fatalf("expected [a b], got %v", results)
	}
}

func TestBroadcast_ClosesOutputs(t *testing.T) {
	in := make(chan int)
	close(in)

	outs := Broadcast(in, 2)
	for i, ch := range outs {
		_, ok := <-ch
		if ok {
			t.Fatalf("expected output channel %d to be closed", i)
		}
	}
}

func TestBroadcast_ConcurrentConsumers(t *testing.T) {
	in := make(chan int)
	const numValues = 50
	const numConsumers = 4

	go func() {
		for i := 0; i < numValues; i++ {
			in <- i
		}
		close(in)
	}()

	outs := Broadcast(in, numConsumers)
	counts := make([]int, numConsumers)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, ch := range outs {
		wg.Add(1)
		go func(idx int, c <-chan int) {
			defer wg.Done()
			count := 0
			for range c {
				count++
			}
			mu.Lock()
			counts[idx] = count
			mu.Unlock()
		}(i, ch)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for broadcast consumers")
	}

	for i, c := range counts {
		if c != numValues {
			t.Fatalf("consumer %d: expected %d values, got %d", i, numValues, c)
		}
	}
}

func TestFanIn_StringType(t *testing.T) {
	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)
	ch1 <- "hello"
	close(ch1)
	ch2 <- "world"
	close(ch2)

	out := FanIn(ch1, ch2)
	var results []string
	for val := range out {
		results = append(results, val)
	}
	sort.Strings(results)
	if len(results) != 2 || results[0] != "hello" || results[1] != "world" {
		t.Fatalf("expected [hello world], got %v", results)
	}
}
