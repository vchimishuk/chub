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
	"sync/atomic"
	"testing"
	"time"

	"github.com/vchimishuk/chub/assert"
)

func TestPeekFree(t *testing.T) {
	nbuf := 16
	c := NewBufferRing(512, nbuf)
	c.Open()
	defer c.Close()

	for i := 0; i < nbuf; i++ {
		b := c.PeekFree()
		assert.True(t, b != nil)
		c.Offer(b)
	}

	b := c.Peek()
	c.OfferFree(b)
	assert.True(t, c.PeekFree() != nil)
}

func TestPeek(t *testing.T) {
	bufsz := 512
	c := NewBufferRing(bufsz, 16)
	c.Open()

	n := byte(127)
	var r atomic.Int32

	go func() {
		for i := byte(0); i < n; i++ {
			b := c.PeekFree()
			assert.True(t, cap(b.data) == bufsz)
			b.data = b.data[0:bufsz]
			b.data[0] = (i % 127)
			b.data[cap(b.data)-1] = -i
			c.Offer(b)
		}

		for r.Load() != int32(n) {
			time.Sleep(time.Millisecond)
		}
		c.Close()
	}()

	go func() {
		for i := byte(0); i < n; i++ {
			b := c.Peek()
			if b == nil {
				break
			}
			assert.True(t, b.data[0] == i)
			assert.True(t, b.data[cap(b.data)-1] == -i)
			r.Store(r.Load() + 1)
			c.OfferFree(b)
		}
	}()

	for r.Load() != int32(n) {
		time.Sleep(time.Millisecond)
	}
	assert.True(t, r.Load() == int32(n))
}
