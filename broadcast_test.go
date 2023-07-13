package broadcast

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnitSignalBroadcast check
func TestUnitSignalBroadcast(t *testing.T) {
	t.Run(`wait`, func(t *testing.T) {
		x := 0
		b := NewSignalBroadcast(&sync.Mutex{})
		done := make(chan bool)
		go func() {
			b.L.Lock()
			x = 1
			b.Wait()
			require.Equal(t, 2, x, "expected 2")
			x = 3
			b.Broadcast()
			b.L.Unlock()
			done <- true
		}()
		go func() {
			b.L.Lock()
			for {
				if x == 1 {
					x = 2
					b.Broadcast()
					break
				}
				b.L.Unlock()
				runtime.Gosched()
				b.L.Lock()
			}
			b.L.Unlock()
			done <- true
		}()
		go func() {
			b.L.Lock()
			for {
				if x == 2 {
					b.Wait()
					require.Equal(t, 3, x, "expected 3")
					break
				}
				if x == 3 {
					break
				}
				b.L.Unlock()
				runtime.Gosched()
				b.L.Lock()
			}
			b.L.Unlock()
			done <- true
		}()
		<-done
		<-done
		<-done
	})
	t.Run(`wait chan`, func(t *testing.T) {
		b := NewSignalBroadcast(&sync.Mutex{})
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
	br := NewSignalBroadcast(&sync.Mutex{})
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
