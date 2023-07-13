package broadcast

import (
	"context"
	"sync"
	"sync/atomic"
	"unsafe"
)

type (
	// SignalBroadcast implementation
	SignalBroadcast struct {
		L sync.Locker
		n unsafe.Pointer
	}
)

// NewSignalBroadcast constructor
func NewSignalBroadcast(l sync.Locker) *SignalBroadcast {
	b := &SignalBroadcast{L: l}
	n := make(chan struct{})
	b.n = unsafe.Pointer(&n)
	return b
}

// Wait for Broadcast() calls
func (b *SignalBroadcast) Wait() {
	n := b.WaitChan()
	b.L.Unlock()
	<-n
	b.L.Lock()
}

// WaitCtx same as Wait() call but with a given context
func (b *SignalBroadcast) WaitCtx(ctx context.Context) {
	n := b.WaitChan()
	b.L.Unlock()
	select {
	case <-n:
	case <-ctx.Done():
	}
	b.L.Lock()
}

// WaitChan returns a channel that can be used to wait for next Broadcast() call
func (b *SignalBroadcast) WaitChan() <-chan struct{} {
	ptr := atomic.LoadPointer(&b.n)
	return *((*chan struct{})(ptr))
}

// Broadcast call notifies all waiters
func (b *SignalBroadcast) Broadcast() {
	n := make(chan struct{})
	ptrOld := atomic.SwapPointer(&b.n, unsafe.Pointer(&n))
	close(*(*chan struct{})(ptrOld))
}
