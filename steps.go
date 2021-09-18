package gobl

import (
	"fmt"
	"os"
	"time"

	"os/exec"
)

// Step is the interface that all gobl steps adhere to.
type Step interface {
	run(Result) chan Result
}

// WatchStep handles setting up watch conditions.
type WatchStep struct {
	Paths []string
}

func (s WatchStep) run(r Result) chan Result {
	result := make(chan Result)
	result <- Result{}
	return result
}

// RunTaskStep handles running a new step.
type RunTaskStep struct {
	TaskName string
}

func (s RunTaskStep) run(r Result) chan Result {
	return RunTask(s.TaskName)
}

// ResultTaskStep handles the result of a previous step.
type ResultTaskStep struct {
	Func func(interface{})
}

func (s ResultTaskStep) run(r Result) chan Result {
	result := make(chan Result)
	go func() {
		s.Func(r.Result)
		result <- Result{nil, nil, nil}
	}()
	return result
}

// CatchTaskStep handles catching errors from any preceding steps.
type CatchTaskStep struct {
	Func func(error) error
}

func (s CatchTaskStep) run(r Result) chan Result {
	result := make(chan Result)
	go func() {
		result <- Result{nil, s.Func(fmt.Errorf("%v: %v", r.Error, r.Result)), nil}
	}()
	return result
}

// EnvStep sets up environment variables to use.
type EnvStep struct {
	Args []string
}

func (s EnvStep) run(pr Result) chan Result {
	result := make(chan Result)
	go func() {
		pr.Task.env = append(pr.Task.env, s.Args...)
		result <- Result{}
	}()
	return result
}

// ExecStep handles executing a command.
type ExecStep struct {
	Args []string
}

func (s ExecStep) run(pr Result) chan Result {
	result := make(chan Result)

	killSignal := make(chan Result)
	doneSignal := make(chan Result)

	pr.Task.processKillChannels = append(pr.Task.processKillChannels, killSignal)

	// Create and set up our command before spawning goroutines
	cmd := exec.Command(s.Args[0], s.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), pr.Task.env...)

	// Loop for either our doneSignal or our external kill signal
	go func() {
		select {
		case <-killSignal:
			if err := cmd.Process.Kill(); err != nil {
				result <- Result{nil, err, nil}
				return
			}
			result <- Result{"killed", nil, nil}
		case r := <-doneSignal:
			result <- r
		}
		// FIXME: This is really overreaching for this step to change the Task's own properties.
		for i, v := range pr.Task.processKillChannels {
			if v == killSignal {
				pr.Task.processKillChannels[i] = pr.Task.processKillChannels[len(pr.Task.processKillChannels)-1]
				pr.Task.processKillChannels = pr.Task.processKillChannels[:len(pr.Task.processKillChannels)-1]
			}
		}
	}()
	// Start and wait for our command.
	go func() {
		if err := cmd.Start(); err != nil {
			doneSignal <- Result{nil, err, nil}
			return
		}
		if err := cmd.Wait(); err != nil {
			doneSignal <- Result{nil, err, nil}
			return
		}
		doneSignal <- Result{nil, nil, nil}
	}()
	return result
}

// ChdirStep handles changing the current working directory.
type ChdirStep struct {
	Path string
}

func (s ChdirStep) run(pr Result) chan Result {
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

// ExistsStep handles checking if a directory or file exists.
type ExistsStep struct {
	Path string
}

func (s ExistsStep) run(pr Result) chan Result {
	result := make(chan Result)

	go func() {
		info, err := os.Stat(s.Path)
		if err != nil {
			result <- Result{nil, err, nil}
			return
		}
		result <- Result{info, nil, nil}
	}()

	return result
}

// SleepStep causes a delay.
type SleepStep struct {
	Duration string
}

func (s SleepStep) run(pr Result) chan Result {
	result := make(chan Result)

	go func() {
		d, err := time.ParseDuration(s.Duration)
		if err != nil {
			result <- Result{nil, err, nil}
			return
		}
		time.Sleep(d)
		result <- Result{s.Duration, nil, nil}
	}()

	return result
}
