package steps

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// ExecStep handles executing a command.
type ExecStep struct {
	Args []interface{}
}

// Run runs a command.
func (s ExecStep) Run(pr Result) chan Result {
	result := make(chan Result)

	killSignal := make(chan Result)
	doneSignal := make(chan Result)

	pr.Context.AddProcessKillChannel(killSignal)

	// Set up buffer for capturing output.
	var buffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &buffer)

	var args []string
	// Convert interface arguments to real arguments.
	for _, a := range s.Args {
		// First dereference pointer types. TODO: Probably replace with reflect.
		var t interface{}
		switch v := a.(type) {
		case *string:
			t = *v
		case *int:
			t = *v
		case *int8:
			t = *v
		case *int16:
			t = *v
		case *int32:
			t = *v
		case *int64:
			t = *v
		case *uint:
			t = *v
		case *uint8:
			t = *v
		case *uint16:
			t = *v
		case *uint32:
			t = *v
		case *uint64:
			t = *v
		case *float32:
			t = *v
		case *float64:
			t = *v
		case *bool:
			t = *v
		default:
			t = v
		}
		// Now convert our argument to a string.
		switch v := t.(type) {
		case string:
			args = append(args, v)
		case bool:
			if v {
				args = append(args, "true")
			} else {
				args = append(args, "false")
			}
		case int:
			args = append(args, fmt.Sprintf("%d", v))
		case int8:
			args = append(args, fmt.Sprintf("%d", v))
		case int16:
			args = append(args, fmt.Sprintf("%d", v))
		case int32:
			args = append(args, fmt.Sprintf("%d", v))
		case int64:
			args = append(args, fmt.Sprintf("%d", v))
		case uint:
			args = append(args, fmt.Sprintf("%d", v))
		case uint8:
			args = append(args, fmt.Sprintf("%d", v))
		case uint16:
			args = append(args, fmt.Sprintf("%d", v))
		case uint32:
			args = append(args, fmt.Sprintf("%d", v))
		case uint64:
			args = append(args, fmt.Sprintf("%d", v))
		case float32:
			args = append(args, fmt.Sprintf("%f", v))
		case float64:
			args = append(args, fmt.Sprintf("%f", v))
		default:
			args = append(args, fmt.Sprintf("%v", v))
		}
	}

	// Create and set up our command before spawning goroutines
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = mw
	cmd.Stderr = os.Stderr
	cmd.Dir = pr.Context.WorkingDirectory()
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
		doneSignal <- Result{buffer.String(), nil, nil}
	}()
	return result
}
