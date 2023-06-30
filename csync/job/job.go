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

type Job interface {
	Shutdown() error
}

type Task func(close <-chan any) error

type impl struct {
	close  chan any
	closed chan error
}

func Start(t Task) Job {
	i := impl{
		close:  make(chan any, 1),
		closed: make(chan error, 1),
	}

	go func() {
		i.closed <- t(i.close)
	}()

	return i
}

func (i impl) Shutdown() error {
	i.close <- struct{}{}

	return <-i.closed
}
