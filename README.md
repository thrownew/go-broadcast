[![Release](https://img.shields.io/github/release/thrownew/go-broadcast.svg)](https://github.com/thrownew/go-broadcast/releases/latest)
[![License](https://img.shields.io/github/license/thrownew/go-broadcast.svg)](https://raw.githubusercontent.com/thrownew/go-broadcast/master/LICENSE)
[![Godocs](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/thrownew/go-broadcast)

[![Build Status](https://github.com/thrownew/go-broadcast/workflows/CI/badge.svg)](https://github.com/thrownew/go-broadcast/actions)
[![codecov](https://codecov.io/gh/thrownew/go-broadcast/branch/master/graph/badge.svg)](https://codecov.io/gh/thrownew/go-broadcast)
[![Go Report Card](https://goreportcard.com/badge/github.com/thrownew/go-broadcast)](https://goreportcard.com/report/github.com/thrownew/go-broadcast)


# go-broadcast

A lightweight and efficient Go package that provides a signal broadcasting mechanism for goroutine synchronization. This implementation offers a unique approach where a single channel is reused across multiple goroutines for each broadcast signal, combining the convenience of channel-based communication with optimal resource usage.

## Features

- Single channel per signal - efficient memory usage while maintaining channel-based synchronization
- Lock-free broadcasting using atomic operations
- Context support for cancellation
- High performance with minimal allocations
- Thread-safe implementation
- Simple API that feels natural to Go developers

## Installation

```bash
go get github.com/thrownew/go-broadcast
```

## Benchmarks

```bash
goos: darwin
goarch: arm64
pkg: github.com/thrownew/go-broadcast
cpu: Apple M1
BenchmarkBroadcastSignalBroadcast
BenchmarkBroadcastSignalBroadcast-8   	12556339	        85.52 ns/op	     120 B/op	       2 allocs/op
BenchmarkBroadcastChannelCopy
BenchmarkBroadcastChannelCopy-8       	 3810374	       314.6 ns/op	       0 B/op	       0 allocs/op
```

## Usage

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"

	"github.com/thrownew/go-broadcast"
)

func main()  {
	b := broadcast.NewSignalBroadcast()

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	// Create 3 goroutines that wait for the signal
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var n int
			for {
				err := b.WaitCtx(ctx)
				if err != nil {
					fmt.Printf("context done: %d %s\n", i, err.Error())
					break
				}
				fmt.Printf("received signal: %d #%d (%s)\n", i, n, time.Now())
				n++
				if n > 10 {
					break
				}
			}
		}(i)
	}

	// Create a goroutine that broadcasts the signal every 100 milliseconds

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fmt.Printf("broadcasting signal (%s)\n", time.Now())
				b.Broadcast()
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
	cancel()
}
```
