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
