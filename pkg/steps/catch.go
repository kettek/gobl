package steps

import "fmt"

// CatchStep handles catching errors from any preceding steps.
type CatchStep struct {
	Func func(error) error
}

// Run runs the catch's function.
func (s CatchStep) Run(r Result) chan Result {
	result := make(chan Result)
	go func() {
		result <- Result{nil, s.Func(fmt.Errorf("%v: %v", r.Error, r.Result)), nil}
	}()
	return result
}
