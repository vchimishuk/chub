// Copyright 2023 Viacheslav Chimishuk <vchimishuk@yandex.ru>
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

package job

import "testing"

func TestShutdownImmediately(t *testing.T) {
	j := Start(func(close <-chan any) error {
		return nil
	})
	err := j.Shutdown()
	if err != nil {
		t.Fatal(err)
	}
}

func TestShutdownWait(t *testing.T) {
	j := Start(func(close <-chan any) error {
		<-close
		return nil
	})
	err := j.Shutdown()
	if err != nil {
		t.Fatal(err)
	}
}
