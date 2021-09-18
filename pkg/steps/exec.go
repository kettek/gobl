package steps

import (
	"os"
	"os/exec"
)

// ExecStep handles executing a command.
type ExecStep struct {
	Args []string
}

// Run runs a command.
func (s ExecStep) Run(pr Result) chan Result {
	result := make(chan Result)

	killSignal := make(chan Result)
	doneSignal := make(chan Result)

	pr.Context.AddProcessKillChannel(killSignal)

	// Create and set up our command before spawning goroutines
	cmd := exec.Command(s.Args[0], s.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = pr.Context.GetEnv()

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
		pr.Context.RemoveProcessKillChannel(killSignal)
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
