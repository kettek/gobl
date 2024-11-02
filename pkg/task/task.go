package task

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/kettek/gobl/pkg/colors"
	"github.com/kettek/gobl/pkg/messages"
	"github.com/kettek/gobl/pkg/steps"

	"github.com/radovskyb/watcher"
)

// Task is a named container for steps.
type Task struct {
	Name           string
	running        bool
	watcher        *watcher.Watcher
	watchPaths     []string
	steps          []steps.Step
	runChannel     chan bool
	stopChannel    chan error
	signalChannels []chan bool
	context        steps.Context
}

// NewTask returns a pointer to a Task with required properties initialized.
func NewTask(name string, context steps.Context) *Task {
	return &Task{
		Name:        name,
		stopChannel: make(chan error),
		runChannel:  make(chan bool),
		watcher:     watcher.New(),
		context:     context,
	}
}

// TODO: We should have each step able to be regular or parallel. Either one would receive an input channel for kill/reset and return an output channel for complete/update/etc.. Parallel steps, such as Watch, would simply run, then processing would immediately continue to the next step (which might block). These parallel operations channels would be added to the Task in a separate slice.
func (g *Task) runSteps() steps.Result {
	// Store working directory so we can restore on close.
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Couldn't get working directory", err)
	}
	g.context.SetWorkingDirectory(wd)
	defer func() {
		if err := os.Chdir(wd); err != nil {
			fmt.Println("Couldn't restore working directory", err)
		}
	}()

	// Variables for Prompt functionality.
	skipToIndex := -1
	queryHandled := false

	prevResult := steps.Result{Result: nil, Error: nil, Context: g.context}
	for i := 0; i < len(g.steps); i++ {
		// skipToIndex is used for "jumping"
		if skipToIndex != -1 {
			i = skipToIndex
			skipToIndex = -1
		}
		step := g.steps[i]

		result := <-step.Run(prevResult)
		result.Context = g.context
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
		} else {
			// Seems safe enough of a spot to put prompt/yes/no handling.
			switch step.(type) {
			case steps.PromptStep:
				queryHandled = false
				yes, nextYesIndex := g.getNextStep(i, steps.YesStep{})
				no, nextNoIndex := g.getNextStep(i, steps.NoStep{})
				_, nextPromptEndIndex := g.getNextStep(i, steps.EndStep{})
				if result.Result == true {
					if yes == nil {
						skipToIndex = nextPromptEndIndex
					} else {
						skipToIndex = nextYesIndex
					}
				} else {
					if no == nil {
						skipToIndex = nextPromptEndIndex
					} else {
						skipToIndex = nextNoIndex
					}
				}
			case steps.YesStep:
				if queryHandled {
					_, nextPromptEndIndex := g.getNextStep(i, steps.EndStep{})
					skipToIndex = nextPromptEndIndex
				}
				queryHandled = true
			case steps.NoStep:
				if queryHandled {
					_, nextPromptEndIndex := g.getNextStep(i, steps.EndStep{})
					skipToIndex = nextPromptEndIndex
				}
				queryHandled = true
			case steps.EndStep:
				queryHandled = false
			}
		}
		prevResult = result
	}
	return prevResult
}

func (g *Task) killProcesses() {
	for _, ch := range g.context.GetProcessKillChannels() {
		ch <- steps.Result{}
	}
}

func (g *Task) getFollowingCatch(pos int) steps.Step {
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

func (g *Task) getNextStep(pos int, target steps.Step) (steps.Step, int) {
	for i := pos + 1; i < len(g.steps); i++ {
		if i >= len(g.steps) {
			return nil, i - 1
		}
		step := g.steps[i]
		if reflect.TypeOf(target) == reflect.TypeOf(step) {
			return step, i
		}
	}
	return nil, pos + 1
}

func (g *Task) runLoop(resultChan chan steps.Result) {
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
			resultChan <- steps.Result{Result: nil, Error: err, Context: g.context}
			g.running = false
			return
		}
	}
}

func (g *Task) watchLoop() {
	if len(g.watcher.WatchedFiles()) > 0 {
		fmt.Printf(messages.WatchingTask+"\n", colors.Info, colors.Clear)
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
							g.kill()
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

func (g *Task) kill() {
	for i := 0; i < len(g.steps); i++ {
		step := g.steps[i]
		switch step := step.(type) {
		case steps.RunStep:
			g2 := GetTask(step.TaskName)
			if g2 != nil && g2.running {
				g2.killProcesses()
			}
		case steps.ParallelStep:
			for _, taskName := range step.TaskNames {
				g2 := GetTask(taskName)
				if g2 != nil && g2.running {
					g2.killProcesses()
				}
			}
		}
	}
}

// Execute runs the given Task.
func (g *Task) Execute() chan steps.Result {
	result := make(chan steps.Result)

	go g.runLoop(result)

	go g.watchLoop()
	return result
}

// Watch sets up a variadic number of glob paths to watch.
func (g *Task) Watch(paths ...string) *Task {
	for _, path := range paths {
		if strings.Contains(path, "**") {
			matches, err := doubleGlob(path)
			if err != nil {
				fmt.Println(err)
			}
			g.watchPaths = append(g.watchPaths, matches...)
		} else {
			matches, err := filepath.Glob(path)
			if err != nil {
				fmt.Println(err)
			}
			g.watchPaths = append(g.watchPaths, matches...)
		}
	}
	for _, file := range g.watchPaths {
		if err := g.watcher.Add(file); err != nil {
			fmt.Println(err)
		}
	}
	return g
}

func doubleGlob(p string) ([]string, error) {
	globs := strings.Split(p, "**")
	if len(globs) == 0 {
		return nil, fmt.Errorf("invalid glob")
	}
	if globs[0] == "" {
		globs[0] = "./"
	}
	matches := make([]string, 1)
	for _, glob := range globs {
		var hits []string
		var hitMap = map[string]bool{}
		for _, match := range matches {
			npath := match + glob
			paths, err := filepath.Glob(npath)
			if err != nil {
				return nil, err
			}
			for _, path := range paths {
				if err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if _, ok := hitMap[path]; !ok {
						hits = append(hits, path)
						hitMap[path] = true
					}
					return nil
				}); err != nil {
					return nil, err
				}
			}
		}
		matches = hits
	}

	return matches, nil
}

// Signaler redirects a given signal to kill steps in the task.
func (g *Task) Signaler(t ...os.Signal) *Task {
	ch := make(chan bool)
	g.signalChannels = append(g.signalChannels, ch)
	sigChan := make(chan os.Signal, 1)
	go func() {
		signal.Notify(sigChan, t...)
		run := true
		for run {
			select {
			case <-ch:
				signal.Reset(t...)
				run = false
				fmt.Println("RESET")
			case <-sigChan:
				g.kill()
				fmt.Println("KILL")
				g.runChannel <- false
			}
		}
	}()
	return g
}

// Catch catches the error of any preceding steps.
func (g *Task) Catch(f func(error) error) *Task {
	g.steps = append(g.steps, steps.CatchStep{
		Func: f,
	})
	return g
}

// Result receives an interface to the result of the last step.
func (g *Task) Result(f func(interface{})) *Task {
	g.steps = append(g.steps, steps.ResultStep{
		Func: f,
	})
	return g
}

// Run runs a task with the given name.
func (g *Task) Run(taskName string) *Task {
	g.steps = append(g.steps, steps.RunStep{
		TaskName: taskName,
	})
	return g
}

// Parallel runs tasks in parallel.
func (g *Task) Parallel(taskNames ...string) *Task {
	g.steps = append(g.steps, steps.ParallelStep{
		TaskNames: taskNames,
	})
	return g
}

// Exec executes a command.
func (g *Task) Exec(args ...interface{}) *Task {
	g.steps = append(g.steps, steps.ExecStep{
		Args: args,
	})
	return g
}

// Env sets environment variables.
func (g *Task) Env(args ...string) *Task {
	g.steps = append(g.steps, steps.EnvStep{
		Args: args,
	})
	return g
}

// Chdir changes the current directory.
func (g *Task) Chdir(path string) *Task {
	g.steps = append(g.steps, steps.ChdirStep{
		Path: path,
	})
	return g
}

// Exists checks if the given file or directory exists.
func (g *Task) Exists(path string) *Task {
	g.steps = append(g.steps, steps.ExistsStep{
		Path: path,
	})
	return g
}

// Sleep delays time by the given string, adhering to https://pkg.go.dev/time#ParseDuration
func (g *Task) Sleep(duration string) *Task {
	g.steps = append(g.steps, steps.SleepStep{
		Duration: duration,
	})
	return g
}

// Print prints either the args passed or the previous task's result.
func (g *Task) Print(args ...interface{}) *Task {
	g.steps = append(g.steps, steps.PrintStep{
		Args: args,
	})
	return g
}

// Prompt prompts (Y/N) using the passed value or the result of the previous task.
func (g *Task) Prompt(v string) *Task {
	g.steps = append(g.steps, steps.PromptStep{
		Message: v,
	})
	return g
}

// End signifies the end of a prompt.
func (g *Task) End() *Task {
	g.steps = append(g.steps, steps.EndStep{})
	return g
}

// Yes signifies the start of steps for a yes prompt result.
func (g *Task) Yes() *Task {
	g.steps = append(g.steps, steps.YesStep{})
	return g
}

// No signifies the start of steps for a no prompt result.
func (g *Task) No() *Task {
	g.steps = append(g.steps, steps.NoStep{})
	return g
}
