package utils

import "sync/atomic"

type AtomicInt struct {
	value int32
}

func NewAtomicInt(i int) AtomicInt {
	return AtomicInt{
		value: int32(i),
	}
}

func (c *AtomicInt) Inc() int {
	return int(atomic.AddInt32((*int32)(&c.value), 1))
}

func (c *AtomicInt) Dec() int {
	return int(atomic.AddInt32((*int32)(&c.value), -1))
}

func (c *AtomicInt) Get() int {
	return int(atomic.LoadInt32((*int32)(&c.value)))
}

func (c *AtomicInt) Set(i int) {
	atomic.StoreInt32((*int32)(&c.value), int32(i))
}

func (c *AtomicInt) Add(i int) int {
	return int(atomic.AddInt32((*int32)(&c.value), int32(i)))
}
