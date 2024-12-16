// Copyright 2016-2024 Viacheslav Chimishuk <vchimishuk@yandex.ru>
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
	"testing"

	"github.com/vchimishuk/chub/assert"
)

func TestOutput(t *testing.T) {
	c, err := Parse(`output = "alsa"`)
	assert.Nil(t, err)
	assert.True(t, c.String("output") == "alsa")

	c, err = Parse(`output = "oss"`)
	assert.Nil(t, err)
	assert.True(t, c.String("output") == "oss")

	_, err = Parse(`output = "unsupported"`)
	assert.Error(t, err, "1: unsupported value")
}

func TestServerHost(t *testing.T) {
	c, err := Parse(`server-host = "localhost"`)
	assert.Nil(t, err)
	assert.True(t, c.String("server-host") == "localhost")
}

func TestServerPort(t *testing.T) {
	c, err := Parse(`server-port = 5115`)
	assert.Nil(t, err)
	assert.True(t, c.Int("server-port") == 5115)
}

func TestVfsRoot(t *testing.T) {
	c, err := Parse(`vfs-root = "/home/user/music"`)
	assert.Nil(t, err)
	assert.True(t, c.String("vfs-root") == "/home/user/music")
}
