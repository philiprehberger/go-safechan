# safechan

[![CI](https://github.com/philiprehberger/go-safechan/actions/workflows/ci.yml/badge.svg)](https://github.com/philiprehberger/go-safechan/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/philiprehberger/go-safechan.svg)](https://pkg.go.dev/github.com/philiprehberger/go-safechan)
[![License](https://img.shields.io/github/license/philiprehberger/go-safechan)](LICENSE)

Safe channel utilities for Go. Provides panic-free send/receive operations, context-aware channel communication, and channel combinators (fan-in, fan-out, broadcast).

## Install

```bash
go get github.com/philiprehberger/go-safechan
```

## Usage

### Safe Send (no panic on closed channel)

```go
import "github.com/philiprehberger/go-safechan"

ch := make(chan int, 1)
ok := safechan.Send(ch, 42) // true

close(ch)
ok = safechan.Send(ch, 1) // false, no panic
```

### Context-aware Send

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()

ch := make(chan int)
ok := safechan.SendCtx(ctx, ch, 42) // false if ctx expires or ch is closed
```

### Context-aware Receive

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()

val, ok := safechan.RecvCtx(ctx, ch) // returns zero value and false on timeout
```

### Fan-In (merge channels)

```go
merged := safechan.FanIn(ch1, ch2, ch3)
for val := range merged {
    fmt.Println(val)
}
```

### Fan-Out (round-robin distribution)

```go
outputs := safechan.FanOut(input, 3)
// Values from input are distributed round-robin to outputs[0], outputs[1], outputs[2]
```

### Broadcast (send to all)

```go
outputs := safechan.Broadcast(input, 3)
// Every value from input is sent to all 3 output channels
```

## API

| Function | Description |
|---|---|
| `Send[T](ch, val)` | Safe send; returns false instead of panicking on closed channel |
| `SendCtx[T](ctx, ch, val)` | Context-aware send; returns false on cancellation or closed channel |
| `Recv[T](ch)` | Receive with ok flag |
| `RecvCtx[T](ctx, ch)` | Context-aware receive; returns zero + false on cancellation |
| `FanIn[T](channels...)` | Merge multiple channels into one |
| `FanOut[T](ch, n)` | Distribute values round-robin to n channels |
| `Broadcast[T](ch, n)` | Send each value to all n channels |

## Development

```bash
go test ./...
go vet ./...
```

## License

MIT
