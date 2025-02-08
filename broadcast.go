package broadcast

import (
	"context"
	"sync/atomic"
	"unsafe"
)

type (
	SignalBroadcast struct {
		n unsafe.Pointer
	}
)

// NewSignalBroadcast constructor
func NewSignalBroadcast() *SignalBroadcast {
	b := &SignalBroadcast{}
	n := make(chan struct{})
	b.n = unsafe.Pointer(&n)
	return b
}

// Wait for Broadcast() calls
func (b *SignalBroadcast) Wait() {
	<-b.WaitChan()
}

// WaitCtx same as Wait() call but with a given context
func (b *SignalBroadcast) WaitCtx(ctx context.Context) error {
	n := b.WaitChan()
	select {
	case <-n:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
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
