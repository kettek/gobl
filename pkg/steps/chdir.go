package steps

import "os"

// ChdirStep handles changing the current working directory.
type ChdirStep struct {
	Path string
}

// Run changes the working directory.
func (s ChdirStep) Run(pr Result) chan Result {
	result := make(chan Result)

	go func() {
		if err := os.Chdir(s.Path); err != nil {
			result <- Result{nil, err, nil}
			return
		}
		wd, err := os.Getwd()
		if err != nil {
			result <- Result{nil, nil, nil}
		}
		result <- Result{wd, nil, nil}
	}()

	return result
}
