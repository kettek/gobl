package steps

import "fmt"

// PrintStep handles printing passed arguments or the results of the previous step if no arguments are passed.
type PrintStep struct {
	Args []interface{}
}

// Run prints the contents of the print step.
func (s PrintStep) Run(r Result) chan Result {
	result := make(chan Result)

	go func() {
		if len(s.Args) == 0 {
			fmt.Println(r.Result)
		} else {
			fmt.Println(s.Args...)
		}
		result <- Result{}
	}()

	return result
}
