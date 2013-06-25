package playingthread

// Playing routine command action.
type action int

const (
	actionPlay action = iota
	actionStop
	actionPause
	actionQuit
)

// Playing routine command.
type command struct {
	// Action to do.
	action action
	// Action arguments if any.
	args []interface{}
	// Channel to write result of command to.
	resultChan chan error
}

// newCommand creates newly initialized command object.
func newCommand(action action, args []interface{}) *command {
	return &command{action: action, args: args, resultChan: make(chan error)}
}
