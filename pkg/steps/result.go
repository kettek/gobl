package steps

// Result represents the result of a step.
type Result struct {
	Result  interface{}
	Error   error
	Context Context
}

// ResultStep handles the result of a previous step.
type ResultStep struct {
	Func func(interface{})
}

// Run calls the step's result function.
func (s ResultStep) Run(r Result) chan Result {
	result := make(chan Result)
	go func() {
		s.Func(r.Result)
		result <- Result{nil, nil, nil}
	}()
	return result
}
