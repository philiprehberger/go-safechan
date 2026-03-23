package safechan

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSend_NormalOperation(t *testing.T) {
	ch := make(chan int, 1)
	ok := Send(ch, 42)
	if !ok {
		t.Fatal("expected Send to return true on open channel")
	}
	val := <-ch
	if val != 42 {
		t.Fatalf("expected 42, got %d", val)
	}
}

func TestSend_ClosedChannel(t *testing.T) {
	ch := make(chan int, 1)
	close(ch)
	ok := Send(ch, 1)
	if ok {
		t.Fatal("expected Send to return false on closed channel")
	}
}

func TestSend_UnbufferedChannel(t *testing.T) {
	ch := make(chan string)
	go func() {
		time.Sleep(10 * time.Millisecond)
		<-ch
	}()
	ok := Send(ch, "hello")
	if !ok {
		t.Fatal("expected Send to return true on unbuffered channel with receiver")
	}
}

func TestSend_ConcurrentSenders(t *testing.T) {
	ch := make(chan int, 100)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			Send(ch, v)
		}(i)
	}
	wg.Wait()
	close(ch)

	count := 0
	for range ch {
		count++
	}
	if count != 100 {
		t.Fatalf("expected 100 values, got %d", count)
	}
}

func TestSendCtx_NormalOperation(t *testing.T) {
	ch := make(chan int, 1)
	ctx := context.Background()
	ok := SendCtx(ctx, ch, 99)
	if !ok {
		t.Fatal("expected SendCtx to return true")
	}
	val := <-ch
	if val != 99 {
		t.Fatalf("expected 99, got %d", val)
	}
}

func TestSendCtx_CancelledContext(t *testing.T) {
	ch := make(chan int) // unbuffered, no receiver
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ok := SendCtx(ctx, ch, 1)
	if ok {
		t.Fatal("expected SendCtx to return false on cancelled context")
	}
}

func TestSendCtx_ClosedChannel(t *testing.T) {
	ch := make(chan int, 1)
	close(ch)
	ctx := context.Background()
	ok := SendCtx(ctx, ch, 1)
	if ok {
		t.Fatal("expected SendCtx to return false on closed channel")
	}
}

func TestSendCtx_ContextTimeout(t *testing.T) {
	ch := make(chan int) // unbuffered, no receiver
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	ok := SendCtx(ctx, ch, 1)
	if ok {
		t.Fatal("expected SendCtx to return false after timeout")
	}
}

func TestRecv_NormalOperation(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 7
	val, ok := Recv(ch)
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if val != 7 {
		t.Fatalf("expected 7, got %d", val)
	}
}

func TestRecv_ClosedChannel(t *testing.T) {
	ch := make(chan int)
	close(ch)
	val, ok := Recv(ch)
	if ok {
		t.Fatal("expected ok to be false on closed channel")
	}
	if val != 0 {
		t.Fatalf("expected zero value, got %d", val)
	}
}

func TestRecv_ClosedChannelWithValues(t *testing.T) {
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	close(ch)

	val, ok := Recv(ch)
	if !ok || val != 1 {
		t.Fatalf("expected (1, true), got (%d, %v)", val, ok)
	}
	val, ok = Recv(ch)
	if !ok || val != 2 {
		t.Fatalf("expected (2, true), got (%d, %v)", val, ok)
	}
	val, ok = Recv(ch)
	if ok {
		t.Fatal("expected ok to be false after draining closed channel")
	}
	if val != 0 {
		t.Fatalf("expected zero value, got %d", val)
	}
}

func TestRecvCtx_NormalOperation(t *testing.T) {
	ch := make(chan string, 1)
	ch <- "test"
	ctx := context.Background()
	val, ok := RecvCtx(ctx, ch)
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if val != "test" {
		t.Fatalf("expected 'test', got %q", val)
	}
}

func TestRecvCtx_CancelledContext(t *testing.T) {
	ch := make(chan int) // unbuffered, no sender
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	val, ok := RecvCtx(ctx, ch)
	if ok {
		t.Fatal("expected ok to be false on cancelled context")
	}
	if val != 0 {
		t.Fatalf("expected zero value, got %d", val)
	}
}

func TestRecvCtx_ClosedChannel(t *testing.T) {
	ch := make(chan int)
	close(ch)
	ctx := context.Background()
	val, ok := RecvCtx(ctx, ch)
	if ok {
		t.Fatal("expected ok to be false on closed channel")
	}
	if val != 0 {
		t.Fatalf("expected zero value, got %d", val)
	}
}

func TestRecvCtx_ContextTimeout(t *testing.T) {
	ch := make(chan int) // unbuffered, no sender
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	val, ok := RecvCtx(ctx, ch)
	if ok {
		t.Fatal("expected ok to be false after timeout")
	}
	if val != 0 {
		t.Fatalf("expected zero value, got %d", val)
	}
}

func TestRecvCtx_WithSender(t *testing.T) {
	ch := make(chan int)
	ctx := context.Background()
	go func() {
		time.Sleep(10 * time.Millisecond)
		ch <- 55
	}()
	val, ok := RecvCtx(ctx, ch)
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if val != 55 {
		t.Fatalf("expected 55, got %d", val)
	}
}

// --- Drain ---

func TestDrain_BufferedChannelWithValues(t *testing.T) {
	ch := make(chan int, 5)
	ch <- 1
	ch <- 2
	ch <- 3
	vals := Drain(ch)
	if len(vals) != 3 {
		t.Fatalf("expected 3 values, got %d", len(vals))
	}
	if vals[0] != 1 || vals[1] != 2 || vals[2] != 3 {
		t.Fatalf("expected [1 2 3], got %v", vals)
	}
}

func TestDrain_EmptyChannel(t *testing.T) {
	ch := make(chan int, 5)
	vals := Drain(ch)
	if len(vals) != 0 {
		t.Fatalf("expected 0 values, got %d", len(vals))
	}
}

func TestDrain_ClosedChannel(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 10
	ch <- 20
	close(ch)
	vals := Drain(ch)
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d", len(vals))
	}
}

func TestDrain_ClosedEmptyChannel(t *testing.T) {
	ch := make(chan int)
	close(ch)
	vals := Drain(ch)
	if len(vals) != 0 {
		t.Fatalf("expected 0 values, got %d", len(vals))
	}
}

// --- DrainCtx ---

func TestDrainCtx_NormalOperation(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	ctx := context.Background()
	vals := DrainCtx(ctx, ch)
	if len(vals) != 3 {
		t.Fatalf("expected 3 values, got %d", len(vals))
	}
}

func TestDrainCtx_CancelledContext(t *testing.T) {
	ch := make(chan int, 5)
	ch <- 1
	ch <- 2
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	vals := DrainCtx(ctx, ch)
	// With a cancelled context, the select may pick either ctx.Done or the channel value.
	// We just verify it doesn't hang and returns at most the buffered values.
	if len(vals) > 2 {
		t.Fatalf("expected at most 2 values, got %d", len(vals))
	}
}

func TestDrainCtx_EmptyChannel(t *testing.T) {
	ch := make(chan int, 5)
	ctx := context.Background()
	vals := DrainCtx(ctx, ch)
	if len(vals) != 0 {
		t.Fatalf("expected 0 values, got %d", len(vals))
	}
}

func TestDrainCtx_ClosedChannel(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 5
	close(ch)
	ctx := context.Background()
	vals := DrainCtx(ctx, ch)
	if len(vals) != 1 {
		t.Fatalf("expected 1 value, got %d", len(vals))
	}
	if vals[0] != 5 {
		t.Fatalf("expected 5, got %d", vals[0])
	}
}

// --- Filter ---

func TestFilter_MatchingValues(t *testing.T) {
	in := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		in <- i
	}
	close(in)

	out := Filter(in, func(v int) bool { return v%2 == 0 })
	var vals []int
	for v := range out {
		vals = append(vals, v)
	}
	if len(vals) != 2 || vals[0] != 2 || vals[1] != 4 {
		t.Fatalf("expected [2 4], got %v", vals)
	}
}

func TestFilter_NoMatches(t *testing.T) {
	in := make(chan int, 3)
	in <- 1
	in <- 3
	in <- 5
	close(in)

	out := Filter(in, func(v int) bool { return v%2 == 0 })
	var vals []int
	for v := range out {
		vals = append(vals, v)
	}
	if len(vals) != 0 {
		t.Fatalf("expected empty, got %v", vals)
	}
}

func TestFilter_ClosedInputClosesOutput(t *testing.T) {
	in := make(chan string)
	close(in)

	out := Filter(in, func(v string) bool { return true })
	_, ok := <-out
	if ok {
		t.Fatal("expected output channel to be closed")
	}
}

func TestFilter_AllMatch(t *testing.T) {
	in := make(chan int, 3)
	in <- 2
	in <- 4
	in <- 6
	close(in)

	out := Filter(in, func(v int) bool { return v%2 == 0 })
	var vals []int
	for v := range out {
		vals = append(vals, v)
	}
	if len(vals) != 3 {
		t.Fatalf("expected 3 values, got %d", len(vals))
	}
}

// --- Map ---

func TestMap_Transform(t *testing.T) {
	in := make(chan int, 3)
	in <- 1
	in <- 2
	in <- 3
	close(in)

	out := Map(in, func(v int) int { return v * 10 })
	var vals []int
	for v := range out {
		vals = append(vals, v)
	}
	if len(vals) != 3 || vals[0] != 10 || vals[1] != 20 || vals[2] != 30 {
		t.Fatalf("expected [10 20 30], got %v", vals)
	}
}

func TestMap_TypeConversion(t *testing.T) {
	in := make(chan int, 3)
	in <- 1
	in <- 2
	in <- 3
	close(in)

	out := Map(in, func(v int) string {
		return string(rune('a' + v - 1))
	})
	var vals []string
	for v := range out {
		vals = append(vals, v)
	}
	if len(vals) != 3 || vals[0] != "a" || vals[1] != "b" || vals[2] != "c" {
		t.Fatalf("expected [a b c], got %v", vals)
	}
}

func TestMap_ClosedInputClosesOutput(t *testing.T) {
	in := make(chan int)
	close(in)

	out := Map(in, func(v int) int { return v })
	_, ok := <-out
	if ok {
		t.Fatal("expected output channel to be closed")
	}
}

func TestMap_EmptyChannel(t *testing.T) {
	in := make(chan int)
	close(in)

	out := Map(in, func(v int) int { return v * 2 })
	var vals []int
	for v := range out {
		vals = append(vals, v)
	}
	if len(vals) != 0 {
		t.Fatalf("expected empty, got %v", vals)
	}
}

// --- SendTimeout ---

func TestSendTimeout_Success(t *testing.T) {
	ch := make(chan int, 1)
	ok := SendTimeout(ch, 42, 100*time.Millisecond)
	if !ok {
		t.Fatal("expected SendTimeout to return true")
	}
	val := <-ch
	if val != 42 {
		t.Fatalf("expected 42, got %d", val)
	}
}

func TestSendTimeout_Timeout(t *testing.T) {
	ch := make(chan int) // unbuffered, no receiver
	ok := SendTimeout(ch, 1, 10*time.Millisecond)
	if ok {
		t.Fatal("expected SendTimeout to return false on timeout")
	}
}

func TestSendTimeout_BufferedFull(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 99 // fill the buffer
	ok := SendTimeout(ch, 1, 10*time.Millisecond)
	if ok {
		t.Fatal("expected SendTimeout to return false on full buffer timeout")
	}
}

// --- RecvTimeout ---

func TestRecvTimeout_Success(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 42
	val, ok := RecvTimeout(ch, 100*time.Millisecond)
	if !ok {
		t.Fatal("expected RecvTimeout to return true")
	}
	if val != 42 {
		t.Fatalf("expected 42, got %d", val)
	}
}

func TestRecvTimeout_Timeout(t *testing.T) {
	ch := make(chan int) // unbuffered, no sender
	val, ok := RecvTimeout(ch, 10*time.Millisecond)
	if ok {
		t.Fatal("expected RecvTimeout to return false on timeout")
	}
	if val != 0 {
		t.Fatalf("expected zero value, got %d", val)
	}
}

func TestRecvTimeout_ClosedChannel(t *testing.T) {
	ch := make(chan int)
	close(ch)
	val, ok := RecvTimeout(ch, 100*time.Millisecond)
	if ok {
		t.Fatal("expected RecvTimeout to return false on closed channel")
	}
	if val != 0 {
		t.Fatalf("expected zero value, got %d", val)
	}
}

func TestRecvTimeout_ClosedChannelWithValues(t *testing.T) {
	ch := make(chan int, 2)
	ch <- 10
	ch <- 20
	close(ch)
	val, ok := RecvTimeout(ch, 100*time.Millisecond)
	if !ok || val != 10 {
		t.Fatalf("expected (10, true), got (%d, %v)", val, ok)
	}
	val, ok = RecvTimeout(ch, 100*time.Millisecond)
	if !ok || val != 20 {
		t.Fatalf("expected (20, true), got (%d, %v)", val, ok)
	}
	val, ok = RecvTimeout(ch, 10*time.Millisecond)
	if ok {
		t.Fatal("expected false after draining closed channel")
	}
}

func TestSend_StringType(t *testing.T) {
	ch := make(chan string, 1)
	ok := Send(ch, "generic")
	if !ok {
		t.Fatal("expected Send to work with string type")
	}
	val := <-ch
	if val != "generic" {
		t.Fatalf("expected 'generic', got %q", val)
	}
}

func TestSend_StructType(t *testing.T) {
	type msg struct {
		ID   int
		Body string
	}
	ch := make(chan msg, 1)
	m := msg{ID: 1, Body: "hello"}
	ok := Send(ch, m)
	if !ok {
		t.Fatal("expected Send to work with struct type")
	}
	val := <-ch
	if val != m {
		t.Fatalf("expected %v, got %v", m, val)
	}
}
