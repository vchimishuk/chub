// Copyright 2024 Viacheslav Chimishuk <vchimishuk@yandex.ru>
//
// This file is part of Chub.
//
// Chub is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Chub is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Chub. If not, see <http://www.gnu.org/licenses/>.

package player

import (
	"sync"

	"github.com/vchimishuk/chub/assert"
)

// Buffer contains piece of decoded PCM data with some metadata attached.
type Buffer struct {
	// Track index in the active playlist.
	plistPos int
	// Current playing track position time.
	trackPos int
	// Audio data.
	data []byte
}

// BufferRing is a ring buffer organized by the list of fixed-size
// reusable buffers. BufferRing is a thread safe, however it is designed
// to work with only one writer and one reader. This schema suits well for
// audio decode-output loop when there is the only one decoder thread and
// one output thread that handles it.
type BufferRing struct {
	// true if BufferRing is open. If BufferRing is closed it is not allowed
	// producer to call Offer() any more.
	open bool
	// Ring array of the buffers.
	bufs []*Buffer
	// Index of the first ready buffer.
	off int
	// Number of consecutive ready buffers started at `off`.
	len int
	// Signals when any buffer becomes free.
	freeCond *sync.Cond
	// Signals when any buffer with data becomes ready for the consumer.
	readyCond *sync.Cond
}

// NewBufferRing creates new BufferRing instance managing `nbuf` buffers of
// `bufsz` bytes each. So the total amount of data which can be stored equals
// to `bufsz * nbuf` bytes.
func NewBufferRing(bufsz int, nbuf int) *BufferRing {
	m := &sync.Mutex{}
	bufs := make([]*Buffer, nbuf)

	for i := 0; i < nbuf; i++ {
		bufs[i] = &Buffer{data: make([]byte, 0, bufsz)}
	}

	return &BufferRing{
		bufs:      bufs,
		freeCond:  sync.NewCond(m),
		readyCond: sync.NewCond(m),
	}
}

// Open opens BufferRing and make it available for the usage.
func (r *BufferRing) Open() {
	r.freeCond.L.Lock()
	defer r.freeCond.L.Unlock()

	r.open = true
	r.off = 0
	r.len = 0
}

// Close closes BufferRing. Consumer is still allowed to consume all data
// contained by BufferRing as usual. After that all next Peek() calls
// will return nil signalling end of the data.
func (r *BufferRing) Close() {
	r.freeCond.L.Lock()
	defer r.freeCond.L.Unlock()

	if !r.open {
		return
	}
	r.open = false
	r.off = 0
	r.len = 0
	r.freeCond.Signal()
	r.readyCond.Signal()
}

// OfferFree marks buffer as free.
// After the buffer is freed it becomes available to be returned by PeekFree().
func (r *BufferRing) OfferFree(b *Buffer) {
	r.freeCond.L.Lock()
	defer r.freeCond.L.Unlock()

	if !r.open {
		return
	}
	assert.True(r.len > 0)

	b.plistPos = 0
	b.trackPos = 0
	b.data = b.data[0:0]
	r.bufs[r.off] = b
	r.off = (r.off + 1) % len(r.bufs)
	r.len--

	r.freeCond.Signal()
}

// PeekFree returns the first empty buffer. If no empty buffer is available
// PeekFree blocks till it will become one. Retruns nil if BufferRing is closed.
func (r *BufferRing) PeekFree() *Buffer {
	r.freeCond.L.Lock()
	defer r.freeCond.L.Unlock()

	if !r.open {
		return nil
	}

	// Wait for a free buffer if ther is no one.
	if r.len == len(r.bufs) {
		r.freeCond.Wait()
		if !r.open {
			return nil
		}
	}

	assert.True(r.len < len(r.bufs))
	i := (r.off + r.len) % len(r.bufs)

	return r.bufs[i]
}

// Offer marks buffer as containing data and available for the consumer.
// Panics if BufferRing is closed.
func (r *BufferRing) Offer(b *Buffer) {
	r.readyCond.L.Lock()
	defer r.readyCond.L.Unlock()

	if !r.open {
		return
	}

	assert.True(r.len < len(r.bufs))
	i := (r.off + r.len) % len(r.bufs)
	r.bufs[i] = b
	r.len++

	r.readyCond.Signal()
}

// Peek returns the first non-emtpy buffer with data. If no buffer is available
// Peek blocks and waits for the one. Returns nil if cache is closed.
func (r *BufferRing) Peek() *Buffer {
	r.readyCond.L.Lock()
	defer r.readyCond.L.Unlock()

	if r.len == 0 {
		if !r.open {
			return nil
		}
		r.readyCond.Wait()
		if r.len == 0 {
			return nil
		}
	}
	assert.True(r.len > 0)

	return r.bufs[r.off]
}
