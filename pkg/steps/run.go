package steps

// RunStep handles Running a new task.
type RunStep struct {
	TaskName string
}

// Run begins running a new task.
func (s RunStep) Run(r Result) chan Result {
	return r.Context.RunTask(s.TaskName)
}
