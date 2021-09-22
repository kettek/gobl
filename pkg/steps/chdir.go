package steps

import (
	"fmt"
	"os"
	"path/filepath"
)

// ChdirStep handles changing the current working directory.
type ChdirStep struct {
	Path string
}

// Run changes the working directory.
func (s ChdirStep) Run(pr Result) chan Result {
	result := make(chan Result)

	go func() {
		wd := filepath.Join(pr.Context.WorkingDirectory(), s.Path)
		info, err := os.Stat(wd)
		if err != nil {
			if os.IsNotExist(err) {
				err = fmt.Errorf("%s does not exist", wd)
			}
			result <- Result{nil, err, nil}
			return
		} else if !info.IsDir() {
			result <- Result{nil, fmt.Errorf("%s is not a directory", wd), nil}
			return
		}
		pr.Context.UpdateWorkingDirectory(wd)
		result <- Result{wd, nil, nil}
	}()

	return result
}
