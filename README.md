# go-safechan

[![CI](https://github.com/philiprehberger/go-safechan/actions/workflows/ci.yml/badge.svg)](https://github.com/philiprehberger/go-safechan/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/philiprehberger/go-safechan.svg)](https://pkg.go.dev/github.com/philiprehberger/go-safechan)
[![Last updated](https://img.shields.io/github/last-commit/philiprehberger/go-safechan)](https://github.com/philiprehberger/go-safechan/commits/main)

Safe channel utilities for Go with panic-free send/receive, context-aware communication, and combinators

## Installation

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

### Drain (collect remaining values)

```go
ch := make(chan int, 10)
ch <- 1
ch <- 2
ch <- 3

vals := safechan.Drain(ch) // [1 2 3], non-blocking

// With context support:
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()
vals = safechan.DrainCtx(ctx, ch) // stops on context cancellation
```

### Filter / Map (channel transformers)

```go
// Filter: keep only even numbers
evens := safechan.Filter(numbers, func(n int) bool { return n%2 == 0 })
for val := range evens {
    fmt.Println(val)
}

// Map: transform values
doubled := safechan.Map(numbers, func(n int) int { return n * 2 })
for val := range doubled {
    fmt.Println(val)
}

// Map with type conversion
strs := safechan.Map(numbers, func(n int) string { return fmt.Sprintf("%d", n) })
```

### Timeout Send / Receive

```go
ch := make(chan int)

// Send with deadline
ok := safechan.SendTimeout(ch, 42, 100*time.Millisecond) // false if timed out

// Receive with deadline
val, ok := safechan.RecvTimeout(ch, 100*time.Millisecond) // zero + false if timed out
```

## API

| Function | Description |
|---|---|
| `Send[T](ch, val)` | Safe send; returns false instead of panicking on closed channel |
| `SendCtx[T](ctx, ch, val)` | Context-aware send; returns false on cancellation or closed channel |
| `Recv[T](ch)` | Receive with ok flag |
| `RecvCtx[T](ctx, ch)` | Context-aware receive; returns zero + false on cancellation |
| `Drain[T](ch)` | Non-blocking read of all buffered values |
| `DrainCtx[T](ctx, ch)` | Context-aware drain; stops on cancellation |
| `Filter[T](ch, pred)` | Forward only values matching predicate |
| `Map[T, R](ch, fn)` | Transform each value through a function |
| `SendTimeout[T](ch, val, d)` | Send with timeout; returns false on expiry |
| `RecvTimeout[T](ch, d)` | Receive with timeout; returns zero + false on expiry |
| `FanIn[T](channels...)` | Merge multiple channels into one |
| `FanOut[T](ch, n)` | Distribute values round-robin to n channels |
| `Broadcast[T](ch, n)` | Send each value to all n channels |

## Development

```bash
go test ./...
go vet ./...
```

## Support

If you find this project useful:

⭐ [Star the repo](https://github.com/philiprehberger/go-safechan)

🐛 [Report issues](https://github.com/philiprehberger/go-safechan/issues?q=is%3Aissue+is%3Aopen+label%3Abug)

💡 [Suggest features](https://github.com/philiprehberger/go-safechan/issues?q=is%3Aissue+is%3Aopen+label%3Aenhancement)

❤️ [Sponsor development](https://github.com/sponsors/philiprehberger)

🌐 [All Open Source Projects](https://philiprehberger.com/open-source-packages)

💻 [GitHub Profile](https://github.com/philiprehberger)

🔗 [LinkedIn Profile](https://www.linkedin.com/in/philiprehberger)

## License

[MIT](LICENSE)
