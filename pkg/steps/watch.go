package steps

// WatchStep handles setting up watch conditions.
type WatchStep struct {
	Paths []string
}

// Run does nothing.
func (s WatchStep) Run(r Result) chan Result {
	result := make(chan Result)
	result <- Result{}
	return result
}
