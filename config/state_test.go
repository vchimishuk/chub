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

package config

import (
	"os"
	"testing"

	"github.com/vchimishuk/chub/assert"
)

func TestLoadState(t *testing.T) {
	f, err := os.CreateTemp("", "chub-tests-*")
	assert.Nil(t, err)
	fname := f.Name()
	f.Close()
	defer os.Remove(fname)

	st := &State{Volume: 45}
	err = SaveState(fname, st)
	assert.Nil(t, err)

	st2, err := LoadState(fname)
	assert.Nil(t, err)
	assert.True(t, st2.Volume == 45)
}
