package bufferpool

import (
	"bytes"
	"sync"
)

// BufferPool serves as a buffer pool of desired bufsize.
type BufferPool struct {
	pool    *sync.Pool
	makebuf func() interface{}
	bufsize int
}

// New returns a newly initialized BufferPool object.
func New(bufsize int) *BufferPool {
	makebuf := func() interface{} {
		return bytes.NewBuffer(make([]byte, bufsize))
	}

	return &BufferPool{
		pool:    &sync.Pool{New: makebuf},
		makebuf: makebuf,
		bufsize: bufsize,
	}
}

// Get selects an arbitrary buffer item from the Pool, removes it from the
// Pool, and returns it to the caller.
func (bp *BufferPool) Get() *bytes.Buffer {
	obj := bp.pool.Get()
	b, ok := obj.(*bytes.Buffer)
	if !ok {
		b = bp.makebuf().(*bytes.Buffer)
	}
	return b
}

// Put resets and adds b to the pool.
func (bp *BufferPool) Put(b *bytes.Buffer) {
	b.Reset()
	bp.pool.Put(b)
}

// BufferSize returns the buffer size of objects in the pool.
func (bp *BufferPool) BufferSize() int {
	return bp.bufsize
}
