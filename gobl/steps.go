package gobl

import (
	"fmt"
	"os"
	"time"

	"os/exec"
)

// GoblStep is the interface that all gobl steps adhere to.
type GoblStep interface {
	run(GoblResult) chan GoblResult
}

// GoblWatchStep handles setting up watch conditions.
type GoblWatchStep struct {
	Paths []string
}

func (s GoblWatchStep) run(r GoblResult) chan GoblResult {
	result := make(chan GoblResult)
	result <- GoblResult{}
	return result
}

// GoblRunTaskStep handles running a new step.
type GoblRunTaskStep struct {
	TaskName string
}

func (s GoblRunTaskStep) run(r GoblResult) chan GoblResult {
	return RunTask(s.TaskName)
}

// GoblResultTaskStep handles the result of a previous step.
type GoblResultTaskStep struct {
	Func func(interface{})
}

func (s GoblResultTaskStep) run(r GoblResult) chan GoblResult {
	result := make(chan GoblResult)
	go func() {
		s.Func(r.Result)
		result <- GoblResult{nil, nil, nil}
	}()
	return result
}

// GoblCatchTaskStep handles catching errors from any preceding steps.
type GoblCatchTaskStep struct {
	Func func(error) error
}

func (s GoblCatchTaskStep) run(r GoblResult) chan GoblResult {
	result := make(chan GoblResult)
	go func() {
		result <- GoblResult{nil, s.Func(fmt.Errorf("%v: %v", r.Error, r.Result)), nil}
	}()
	return result
}

// GoblEnvStep sets up environment variables to use.
type GoblEnvStep struct {
	Args []string
}

func (s GoblEnvStep) run(pr GoblResult) chan GoblResult {
	result := make(chan GoblResult)
	go func() {
		pr.Task.env = append(pr.Task.env, s.Args...)
		result <- GoblResult{}
	}()
	return result
}

// GoblExecStep handles executing a command.
type GoblExecStep struct {
	Args []string
}

func (s GoblExecStep) run(pr GoblResult) chan GoblResult {
	result := make(chan GoblResult)

	killSignal := make(chan GoblResult)
	doneSignal := make(chan GoblResult)

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
				result <- GoblResult{nil, err, nil}
				return
			}
			result <- GoblResult{"killed", nil, nil}
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
			doneSignal <- GoblResult{nil, err, nil}
			return
		}
		if err := cmd.Wait(); err != nil {
			doneSignal <- GoblResult{nil, err, nil}
			return
		}
		doneSignal <- GoblResult{nil, nil, nil}
	}()
	return result
}

// GoblChdirStep handles changing the current working directory.
type GoblChdirStep struct {
	Path string
}

func (s GoblChdirStep) run(pr GoblResult) chan GoblResult {
	result := make(chan GoblResult)

	go func() {
		if err := os.Chdir(s.Path); err != nil {
			result <- GoblResult{nil, err, nil}
			return
		}
		wd, err := os.Getwd()
		if err != nil {
			result <- GoblResult{nil, nil, nil}
		}
		result <- GoblResult{wd, nil, nil}
	}()

	return result
}

// GoblExistsStep handles checking if a directory or file exists.
type GoblExistsStep struct {
	Path string
}

func (s GoblExistsStep) run(pr GoblResult) chan GoblResult {
	result := make(chan GoblResult)

	go func() {
		info, err := os.Stat(s.Path)
		if err != nil {
			result <- GoblResult{nil, err, nil}
			return
		}
		result <- GoblResult{info, nil, nil}
	}()

	return result
}

// GoblSleepStep causes a delay.
type GoblSleepStep struct {
	Duration string
}

func (s GoblSleepStep) run(pr GoblResult) chan GoblResult {
	result := make(chan GoblResult)

	go func() {
		d, err := time.ParseDuration(s.Duration)
		if err != nil {
			result <- GoblResult{nil, err, nil}
			return
		}
		time.Sleep(d)
		result <- GoblResult{s.Duration, nil, nil}
	}()

	return result
}

// GoblCanKill does nothing.
type GoblCanKill struct {
}
