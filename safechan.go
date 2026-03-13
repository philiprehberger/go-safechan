// Package safechan provides safe channel utilities for Go.
//
// It offers panic-free send operations, context-aware channel communication,
// and channel combinators such as fan-in, fan-out, and broadcast.
package safechan

import "context"

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
