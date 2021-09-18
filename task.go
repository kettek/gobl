package gobl

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kettek/gobl/pkg/steps"

	"github.com/radovskyb/watcher"
)

// GoblTask is a named container for steps.
type GoblTask struct {
	Name        string
	running     bool
	watcher     *watcher.Watcher
	watchPaths  []string
	steps       []steps.Step
	runChannel  chan bool
	stopChannel chan error
	context     Context
}

// TODO: We should have each step able to be regular or parallel. Either one would receive an input channel for kill/reset and return an output channel for complete/update/etc.. Parallel steps, such as Watch, would simply run, then processing would immediately continue to the next step (which might block). These parallel operations channels would be added to the Task in a separate slice.
func (g *GoblTask) runSteps() steps.Result {
	// Store working directory so we can restore on close.
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Couldn't get working directory", err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			fmt.Println("Couldn't restore working directory", err)
		}
	}()

	prevResult := steps.Result{Context: &g.context}
	for i := 0; i < len(g.steps); i++ {
		step := g.steps[i]

		result := <-step.Run(prevResult)
		result.Context = &g.context
		catchStep := g.getFollowingCatch(i)
		if result.Error != nil {
			if catchStep == nil {
				return result
			}
			if catchResult := <-catchStep.Run(result); catchResult.Error != nil {
				return catchResult
			}
		}
		if catchStep != nil {
			i++
		}
		prevResult = result
	}
	return prevResult
}

func (g *GoblTask) killProcesses() {
	for _, ch := range g.context.processKillChannels {
		ch <- steps.Result{}
	}
}

func (g *GoblTask) getFollowingCatch(pos int) steps.Step {
	if pos+1 >= len(g.steps) {
		return nil
	}
	step := g.steps[pos+1]
	switch step := step.(type) {
	case steps.CatchStep:
		return &step
	}
	return nil
}

/*func (g *GoblTask) getFollowingsteps.Result(pos int) *steps.ResultStep {
	if pos+1 >= len(g.steps) {
		return nil
	}
	step := g.steps[pos+1]
	switch step := step.(type) {
	case steps.ResultStep:
		return &step
	}
	return nil
}*/

func (g *GoblTask) runLoop(resultChan chan steps.Result) {
	g.running = true
	for {
		select {
		case shouldExit := <-g.runChannel:
			result := g.runSteps()
			if shouldExit {
				resultChan <- result
				g.running = false
				return
			}
		case err := <-g.stopChannel:
			resultChan <- steps.Result{Result: nil, Error: err, Context: &g.context}
			g.running = false
			return
		}
	}
}

func (g *GoblTask) watchLoop() {
	if len(g.watcher.WatchedFiles()) > 0 {
		fmt.Printf("👀  %sWatching%s\n", InfoColor, Clear)
		for k := range g.watcher.WatchedFiles() {
			fmt.Printf("\t%s\n", k)
		}
		// Watch events goroutine.
		go func() {
			g.runChannel <- false // Initial run
			for {
				select {
				case <-g.watcher.Event:
					// Yeah, I know this is racey.
					if g.running {
						if len(g.runChannel) == 0 {
							for i := 0; i < len(g.steps); i++ {
								step := g.steps[i]
								switch step := step.(type) {
								case steps.RunStep:
									g2 := Tasks[step.TaskName]
									if g2.running {
										g2.killProcesses()
									}
								}
							}
							g.runChannel <- false
						}
					}
				case err := <-g.watcher.Error:
					g.stopChannel <- err
				case <-g.watcher.Closed:
					g.stopChannel <- nil
					return
				}
			}
		}()

		// Watch goroutine.
		go func() {
			if err := g.watcher.Start(time.Millisecond * 100); err != nil {
				g.watcher.Close()
			}
		}()
	} else {
		g.runChannel <- true
	}
}

func (g *GoblTask) run() chan steps.Result {
	result := make(chan steps.Result)

	go g.runLoop(result)

	go g.watchLoop()
	return result
}

// Watch sets up a variadic number of glob paths to watch.
func (g *GoblTask) Watch(paths ...string) *GoblTask {
	for _, path := range paths {
		matches, err := filepath.Glob(path)
		if err != nil {
			fmt.Println(err)
		}
		g.watchPaths = append(g.watchPaths, matches...)

		for _, file := range g.watchPaths {
			if err := g.watcher.Add(file); err != nil {
				fmt.Println(err)
			}
		}
	}
	return g
}

// Catch catches the error of any preceding steps.
func (g *GoblTask) Catch(f func(error) error) *GoblTask {
	g.steps = append(g.steps, steps.CatchStep{
		Func: f,
	})
	return g
}

// Result receives an interface to the result of the last step.
func (g *GoblTask) Result(f func(interface{})) *GoblTask {
	g.steps = append(g.steps, steps.ResultStep{
		Func: f,
	})
	return g
}

// Run runs a task with the given name.
func (g *GoblTask) Run(taskName string) *GoblTask {
	g.steps = append(g.steps, steps.RunStep{
		TaskName: taskName,
	})
	return g
}

// Exec executes a command.
func (g *GoblTask) Exec(args ...string) *GoblTask {
	g.steps = append(g.steps, steps.ExecStep{
		Args: args,
	})
	return g
}

// Env sets environment variables.
func (g *GoblTask) Env(args ...string) *GoblTask {
	g.steps = append(g.steps, steps.EnvStep{
		Args: args,
	})
	return g
}

// Chdir changes the current directory.
func (g *GoblTask) Chdir(path string) *GoblTask {
	g.steps = append(g.steps, steps.ChdirStep{
		Path: path,
	})
	return g
}

// Exists checks if the given file or directory exists.
func (g *GoblTask) Exists(path string) *GoblTask {
	g.steps = append(g.steps, steps.ExistsStep{
		Path: path,
	})
	return g
}

// Sleep delays time by the given string, adhering to https://pkg.go.dev/time#ParseDuration
func (g *GoblTask) Sleep(duration string) *GoblTask {
	g.steps = append(g.steps, steps.SleepStep{
		Duration: duration,
	})
	return g
}
