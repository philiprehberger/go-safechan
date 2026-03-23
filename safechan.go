// Package safechan provides safe channel utilities for Go.
//
// It offers panic-free send operations, context-aware channel communication,
// and channel combinators such as fan-in, fan-out, and broadcast.
package safechan

import (
	"context"
	"time"
)

// Send sends val on ch without panicking if ch is closed.
// It returns true if the value was sent successfully, or false if the channel
// was closed (recovering the panic internally).
func Send[T any](ch chan<- T, val T) (sent bool) {
	defer func() {
		if r := recover(); r != nil {
			sent = false
		}
	}()
	ch <- val
	return true
}

// SendCtx sends val on ch, respecting context cancellation.
// It returns true if the value was sent successfully, or false if the context
// was done or the channel was closed.
func SendCtx[T any](ctx context.Context, ch chan<- T, val T) bool {
	defer func() {
		recover()
	}()
	select {
	case <-ctx.Done():
		return false
	case ch <- val:
		return true
	}
}

// Recv receives a value from ch.
// It returns the value and true if a value was received, or the zero value
// and false if the channel is closed.
// This is a thin wrapper around the built-in receive for API consistency.
func Recv[T any](ch <-chan T) (val T, ok bool) {
	val, ok = <-ch
	return val, ok
}

// RecvCtx receives a value from ch, respecting context cancellation.
// It returns the value and true if a value was received, or the zero value
// and false if the context was done or the channel is closed.
func RecvCtx[T any](ctx context.Context, ch <-chan T) (val T, ok bool) {
	select {
	case <-ctx.Done():
		var zero T
		return zero, false
	case val, ok = <-ch:
		return val, ok
	}
}

// Drain reads all remaining values from ch without blocking.
// It returns the collected values immediately when no more values are
// available in the channel buffer.
func Drain[T any](ch <-chan T) []T {
	var result []T
	for {
		select {
		case val, ok := <-ch:
			if !ok {
				return result
			}
			result = append(result, val)
		default:
			return result
		}
	}
}

// DrainCtx reads all remaining values from ch, stopping when the context
// is cancelled or no more values are available in the channel buffer.
func DrainCtx[T any](ctx context.Context, ch <-chan T) []T {
	var result []T
	for {
		select {
		case <-ctx.Done():
			return result
		case val, ok := <-ch:
			if !ok {
				return result
			}
			result = append(result, val)
		default:
			return result
		}
	}
}

// Filter returns a new channel that only forwards values from in that
// satisfy the predicate. A background goroutine reads from in and writes
// matching values to the output channel. The output channel is closed
// when the input channel is closed.
func Filter[T any](in <-chan T, pred func(T) bool) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for val := range in {
			if pred(val) {
				out <- val
			}
		}
	}()
	return out
}

// Map returns a new channel that transforms each value from in using fn.
// A background goroutine reads from in, applies fn, and writes the result
// to the output channel. The output channel is closed when the input
// channel is closed.
func Map[T, R any](in <-chan T, fn func(T) R) <-chan R {
	out := make(chan R)
	go func() {
		defer close(out)
		for val := range in {
			out <- fn(val)
		}
	}()
	return out
}

// SendTimeout sends val on ch with a timeout duration.
// It returns true if the value was sent successfully, or false if the
// deadline expired before the send could complete.
func SendTimeout[T any](ch chan<- T, val T, d time.Duration) bool {
	select {
	case ch <- val:
		return true
	case <-time.After(d):
		return false
	}
}

// RecvTimeout receives a value from ch with a timeout duration.
// It returns the value and true if a value was received, or the zero value
// and false if the deadline expired before a value was available.
func RecvTimeout[T any](ch <-chan T, d time.Duration) (T, bool) {
	select {
	case val, ok := <-ch:
		return val, ok
	case <-time.After(d):
		var zero T
		return zero, false
	}
}
