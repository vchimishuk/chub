// Copyright 2016 Viacheslav Chimishuk <vchimishuk@yandex.ru>
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
// along with Chub.  If not, see <http://www.gnu.org/licenses/>.

package player

type command int

const (
	cmdNext command = iota
	cmdPause
	cmdPlay
	cmdPlist
	cmdPrev
	cmdQuit
	cmdStop
)

type message struct {
	cmd  command
	args []interface{}
	res  chan error
}

func newMessage(cmd command, args []interface{}) *message {
	return &message{cmd: cmd, args: args, res: make(chan error, 1)}
}

func (msg *message) SendResult(err error) {
	msg.res <- err
}

func (msg *message) GetResult() error {
	return <-msg.res
}
