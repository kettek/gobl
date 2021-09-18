package gobl

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/radovskyb/watcher"
)

// GoblTask is a named container for steps.
type GoblTask struct {
	Name                string
	running             bool
	env                 []string
	watcher             *watcher.Watcher
	watchPaths          []string
	steps               []Step
	runChannel          chan bool
	stopChannel         chan error
	processKillChannels []chan Result
}

func (g *GoblTask) runSteps() Result {
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

	prevResult := Result{Task: g}
	for i := 0; i < len(g.steps); i++ {
		step := g.steps[i]

		goblResult := <-step.run(prevResult)
		goblResult.Task = g
		catchStep := g.getFollowingCatch(i)
		if goblResult.Error != nil {
			if catchStep == nil {
				return goblResult
			}
			if catchResult := <-catchStep.run(goblResult); catchResult.Error != nil {
				return catchResult
			}
		}
		if catchStep != nil {
			i++
		}
		prevResult = goblResult
	}
	return prevResult
}

func (g *GoblTask) killProcesses() {
	for _, ch := range g.processKillChannels {
		ch <- Result{}
	}
}

func (g *GoblTask) getFollowingCatch(pos int) *CatchTaskStep {
	if pos+1 >= len(g.steps) {
		return nil
	}
	step := g.steps[pos+1]
	switch step := step.(type) {
	case CatchTaskStep:
		return &step
	}
	return nil
}

/*func (g *GoblTask) getFollowingResult(pos int) *ResultTaskStep {
	if pos+1 >= len(g.steps) {
		return nil
	}
	step := g.steps[pos+1]
	switch step := step.(type) {
	case ResultTaskStep:
		return &step
	}
	return nil
}*/

func (g *GoblTask) runLoop(resultChan chan Result) {
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
			resultChan <- Result{nil, err, g}
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
								case RunTaskStep:
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

func (g *GoblTask) run() chan Result {
	result := make(chan Result)

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
	g.steps = append(g.steps, CatchTaskStep{
		Func: f,
	})
	return g
}

// Result receives an interface to the result of the last step.
func (g *GoblTask) Result(f func(interface{})) *GoblTask {
	g.steps = append(g.steps, ResultTaskStep{
		Func: f,
	})
	return g
}

// Run runs a task with the given name.
func (g *GoblTask) Run(taskName string) *GoblTask {
	g.steps = append(g.steps, RunTaskStep{
		TaskName: taskName,
	})
	return g
}

// Exec executes a command.
func (g *GoblTask) Exec(args ...string) *GoblTask {
	g.steps = append(g.steps, ExecStep{
		Args: args,
	})
	return g
}

// Env sets environment variables.
func (g *GoblTask) Env(args ...string) *GoblTask {
	g.steps = append(g.steps, EnvStep{
		Args: args,
	})
	return g
}

// Chdir changes the current directory.
func (g *GoblTask) Chdir(path string) *GoblTask {
	g.steps = append(g.steps, ChdirStep{
		Path: path,
	})
	return g
}

// Exists checks if the given file or directory exists.
func (g *GoblTask) Exists(path string) *GoblTask {
	g.steps = append(g.steps, ExistsStep{
		Path: path,
	})
	return g
}

// Sleep delays time by the given string, adhering to https://pkg.go.dev/time#ParseDuration
func (g *GoblTask) Sleep(duration string) *GoblTask {
	g.steps = append(g.steps, SleepStep{
		Duration: duration,
	})
	return g
}
