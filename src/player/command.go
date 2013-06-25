// player package is the core of the program: it manages playlists and player's state.
package player

type commandHandler func(args ...interface{}) *result

// Command for communication with player's command routine.
type command struct {
	// Method to invoke with args.
	method commandHandler
	// Args for the method.
	args []interface{}
	// Command handling routine returns response through this channel.
	resultChan chan *result
}

// Command routine's result object.
type result struct {
	// Command processing result itself.
	args []interface{}
	// Error if any.
	err error
}

// newCommand creates newly initialized command object.
func newCommand(method commandHandler, args ...interface{}) *command {
	return &command{method: method, args: args, resultChan: make(chan *result)}
}

// newEmptyResult returns result object with no error and parameters.
func newEmptyResult() *result {
	return &result{}
}

// newErrorResult returns error result.
func newErrorResult(err error) *result {
	return &result{err: err}
}

// newResult returns result object with some arguments.
func newResult(args ...interface{}) *result {
	return &result{args: args}
}
