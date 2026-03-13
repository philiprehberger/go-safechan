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
