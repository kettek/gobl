package steps

// EnvStep sets up environment variables to use.
type EnvStep struct {
	Args []string
}

// Run adds the env var to the task.
func (s EnvStep) Run(pr Result) chan Result {
	result := make(chan Result)
	go func() {
		pr.Context.AddEnv(s.Args...)
		result <- Result{}
	}()
	return result
}
