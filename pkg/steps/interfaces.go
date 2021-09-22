package steps

// Context is an interface to Context.
type Context interface {
	AddProcessKillChannel(chan Result)
	RemoveProcessKillChannel(chan Result)
	GetProcessKillChannels() []chan Result
	GetEnv() []string
	AddEnv(...string)
	RunTask(string) chan Result
	WorkingDirectory() string
	SetWorkingDirectory(string)
	UpdateWorkingDirectory(string)
}

// Step is the interface that all gobl steps adhere to.
type Step interface {
	Run(Result) chan Result
}
