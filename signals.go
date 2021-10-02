package gobl

import "syscall"

// Our various Signals
const (
	SigQuit      = syscall.SIGQUIT
	SigInterrupt = syscall.SIGINT
)
