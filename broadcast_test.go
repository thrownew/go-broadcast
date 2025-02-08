package broadcast

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUnitSignalBroadcast check
func TestUnitSignalBroadcast(t *testing.T) {
	t.Run(`wait`, func(_ *testing.T) {
		var x atomic.Int32
		b := NewSignalBroadcast()
		var wg sync.WaitGroup
		wg.Add(4)
		go func() {
			defer wg.Done()
			for {
				if x.CompareAndSwap(0, 1) {
					b.Wait()
					break
				}
				runtime.Gosched()
			}
		}()
		go func() {
			defer wg.Done()
			for {
				if x.CompareAndSwap(1, 2) {
					b.Wait()
					break
				}
				runtime.Gosched()
			}
		}()
		go func() {
			defer wg.Done()
			for {
				if x.CompareAndSwap(2, 3) {
					b.Broadcast()
					break
				}
				runtime.Gosched()
			}
		}()
		go func() {
			defer wg.Done()
			for {
				if x.CompareAndSwap(3, 4) {
					b.Broadcast()
					break
				}
				runtime.Gosched()
			}
		}()
		wg.Wait()
	})
	t.Run(`wait chan`, func(t *testing.T) {
		b := NewSignalBroadcast()
		var ready, done sync.WaitGroup
		var cnt int32
		for i := 0; i < 8; i++ {
			done.Add(1)
			ready.Add(1)
			go func() {
				defer done.Done()
				ready.Done()
				<-b.WaitChan()
				atomic.AddInt32(&cnt, 1)
			}()
		}
		b.Broadcast()
		ready.Wait()
		b.Broadcast()
		b.Broadcast()
		done.Wait()
		assert.Equal(t, int32(8), cnt)
	})
}

// BenchmarkBroadcastSignalBroadcast check
func BenchmarkBroadcastSignalBroadcast(b *testing.B) {
	b.ReportAllocs()
	br := NewSignalBroadcast()
	stop := make(chan struct{})
	var ready, done sync.WaitGroup
	for i := 0; i < 100; i++ {
		ready.Add(1)
		done.Add(1)
		go func() {
			defer done.Done()
			ready.Done()
			select {
			case <-br.WaitChan():
			case <-stop:
			}
		}()
	}
	ready.Wait()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			br.Broadcast()
		}
	})
	b.StopTimer()
	close(stop)
	done.Wait()
}

// BenchmarkBroadcastChannelCopy check
func BenchmarkBroadcastChannelCopy(b *testing.B) {
	b.ReportAllocs()
	var mu sync.Mutex
	channels := make([]chan struct{}, 0, 100)
	stop := make(chan struct{})
	var ready, done sync.WaitGroup
	for i := 0; i < 100; i++ {
		ch := make(chan struct{})
		channels = append(channels, ch)
		ready.Add(1)
		done.Add(1)
		go func(ch <-chan struct{}) {
			defer done.Done()
			ready.Done()
			select {
			case <-ch:
			case <-stop:
			}
		}(ch)
	}
	ready.Wait()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			for _, ch := range channels {
				select {
				case ch <- struct{}{}:
				default:
				}
			}
			mu.Unlock()
		}
	})
	b.StopTimer()
	close(stop)
	done.Wait()
}
