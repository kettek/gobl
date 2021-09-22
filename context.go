package gobl

import (
	"os"
	"path/filepath"

	"github.com/kettek/gobl/pkg/steps"
)

// Context provides a task-specific set of properties.
type Context struct {
	env                      []string
	processKillChannels      []chan steps.Result
	workingDirectory         string
	originalWorkingDirectory string
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

// GetProcessKillChannels returns the underlying process kill channels slice.
func (c *Context) GetProcessKillChannels() []chan steps.Result {
	return c.processKillChannels
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

// WorkingDirectory returns the context's working directory.
func (c *Context) WorkingDirectory() string {
	return c.workingDirectory
}

// SetWorkingDirectory sets the context's working directory, including settings its original working directory.
func (c *Context) SetWorkingDirectory(wd string) {
	c.workingDirectory, _ = filepath.Abs(wd)
	c.originalWorkingDirectory = c.workingDirectory
}

// UpdateWorkingDirectory updates the context's working directory.
func (c *Context) UpdateWorkingDirectory(wd string) {
	c.workingDirectory = wd
}
