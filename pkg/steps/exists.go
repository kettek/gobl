package steps

import (
	"os"
	"path/filepath"
)

// ExistsStep handles checking if a directory or file exists.
type ExistsStep struct {
	Path string
}

// Run checks if the file or folder exists and returns an fs.FileInfo.
func (s ExistsStep) Run(pr Result) chan Result {
	result := make(chan Result)

	go func() {
		info, err := os.Stat(filepath.Join(pr.Context.WorkingDirectory(), s.Path))
		if err != nil {
			result <- Result{nil, err, nil}
			return
		}
		result <- Result{info, nil, nil}
	}()

	return result
}
