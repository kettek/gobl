package gobl

import (
	"fmt"
	"os"
	"os/exec"
)

type GoblStep interface {
	run(GoblResult) chan GoblResult
}
type GoblWatchStep struct {
	Path string
}

func (s GoblWatchStep) run(r GoblResult) chan GoblResult {
	result := make(chan GoblResult)
	result <- GoblResult{}
	return result
}

type GoblRunTaskStep struct {
	TaskName string
}

func (s GoblRunTaskStep) run(r GoblResult) chan GoblResult {
	return RunTask(s.TaskName)
}

type GoblResultTaskStep struct {
	Func func(interface{})
}

func (s GoblResultTaskStep) run(r GoblResult) chan GoblResult {
	result := make(chan GoblResult)
	go func() {
		s.Func(r.Result)
		result <- GoblResult{nil, nil}
	}()
	return result
}

type GoblCatchTaskStep struct {
	Func func(error) error
}

func (s GoblCatchTaskStep) run(r GoblResult) chan GoblResult {
	result := make(chan GoblResult)
	go func() {
		result <- GoblResult{nil, s.Func(fmt.Errorf("%v: %v", r.Error, r.Result))}
	}()
	return result
}

type GoblExecStep struct {
	Args       []string
	killSignal chan GoblResult
}

func (s GoblExecStep) run(pr GoblResult) chan GoblResult {
	result := make(chan GoblResult)

	s.killSignal = make(chan GoblResult)
	doneSignal := make(chan GoblResult)

	// Create and set up our command before spawning goroutines
	cmd := exec.Command(s.Args[0], s.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Loop for either our doneSignal or our external kill signal
	go func() {
		select {
		case r := <-doneSignal:
			result <- r
		case <-s.killSignal:
			if err := cmd.Process.Kill(); err != nil {
				result <- GoblResult{nil, err}
				return
			}
			result <- GoblResult{"killed", nil}
		}
	}()
	// Start and wait for our command.
	go func() {
		if err := cmd.Start(); err != nil {
			doneSignal <- GoblResult{nil, err}
			return
		}
		if err := cmd.Wait(); err != nil {
			doneSignal <- GoblResult{nil, err}
			return
		}
		doneSignal <- GoblResult{nil, nil}
	}()
	return result
}

type GoblCanKill struct {
}
