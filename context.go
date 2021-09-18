package gobl

import (
	"os"

	"github.com/kettek/gobl/pkg/steps"
)

// Context provides a task-specific set of properties.
type Context struct {
	env                 []string
	processKillChannels []chan steps.Result
}

// AddProcessKillChannel adds a provided channel to be sent a killed signal.
func (c *Context) AddProcessKillChannel(r chan steps.Result) {
	c.processKillChannels = append(c.processKillChannels, r)
}

// RemoveProcessKillChannel removes a provided channel from the process kill slice.
func (c *Context) RemoveProcessKillChannel(r chan steps.Result) {
	for i, v := range c.processKillChannels {
		if v == r {
			c.processKillChannels[i] = c.processKillChannels[len(c.processKillChannels)-1]
			c.processKillChannels = c.processKillChannels[:len(c.processKillChannels)-1]
		}
	}
}

// GetEnv returns the current environment variables, including OS.
func (c *Context) GetEnv() []string {
	return append(os.Environ(), c.env...)
}

// AddEnv adds an environment variable.
func (c *Context) AddEnv(args ...string) {
	c.env = append(c.env, args...)
}

// RunTask runs a task.
func (c *Context) RunTask(n string) chan steps.Result {
	return RunTask(n)
}
