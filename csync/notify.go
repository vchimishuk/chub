// Copyright 2019 Viacheslav Chimishuk <vchimishuk@yandex.ru>
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

package csync

type Msg struct {
	Data   interface{}
	Result chan interface{}
}

type Notify struct {
	ch chan *Msg
}

func NewNotify() *Notify {
	return &Notify{make(chan *Msg)}
}

func (n *Notify) Wait() *Msg {
	return <-n.WaitChan()
}

func (n *Notify) WaitChan() <-chan *Msg {
	return n.ch
}

func (n *Notify) Send(msg interface{}) <-chan interface{} {
	m := &Msg{Data: msg, Result: make(chan interface{})}
	go func() {
		n.ch <- m
	}()

	return m.Result
}
