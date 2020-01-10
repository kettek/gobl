package gobl

import (
	"bytes"
	"fmt"
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
	Args []string
}

func (s GoblExecStep) run(pr GoblResult) chan GoblResult {
	result := make(chan GoblResult)
	go func() {
		cmd := exec.Command(s.Args[0], s.Args[1:]...)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Start(); err != nil {
			result <- GoblResult{nil, err}
			return
		}
		if err := cmd.Wait(); err != nil {
			result <- GoblResult{out.String(), err}
			return
		}
		result <- GoblResult{out.String(), nil}
	}()
	return result
}

type GoblCanKill struct {
}
