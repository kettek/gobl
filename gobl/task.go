package gobl

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/radovskyb/watcher"
)

type GoblResult struct {
	Result interface{}
	Error  error
}

var goblTasks = make(map[string]*GoblTask)

type GoblTask struct {
	Name        string
	watcher     *watcher.Watcher
	watchPaths  []string
	steps       []GoblStep
	stepIndex   int
	channel     chan GoblStep
	runChannel  chan bool
	stopChannel chan error
}

func (g *GoblTask) runSteps() GoblResult {
	prevResult := GoblResult{}
	for i := 0; i < len(g.steps); i++ {
		step := g.steps[i]

		goblResult := <-step.run(prevResult)
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

func (g *GoblTask) getFollowingCatch(pos int) *GoblCatchTaskStep {
	if pos+1 >= len(g.steps) {
		return nil
	}
	step := g.steps[pos+1]
	switch step := step.(type) {
	case GoblCatchTaskStep:
		return &step
	}
	return nil
}

func (g *GoblTask) getFollowingResult(pos int) *GoblResultTaskStep {
	if pos+1 >= len(g.steps) {
		return nil
	}
	step := g.steps[pos+1]
	switch step := step.(type) {
	case GoblResultTaskStep:
		return &step
	}
	return nil
}

func (g *GoblTask) getNextStep() GoblStep {
	if g.stepIndex+1 >= len(g.steps) {
		return nil
	}
	return g.steps[g.stepIndex+1]
}

func (g *GoblTask) runNextStep() GoblResult {
	if g.stepIndex+1 >= len(g.steps) {
		g.stepIndex = 0
		return GoblResult{nil, nil}
	}
	g.stepIndex++
	switch step := g.steps[g.stepIndex].(type) {
	case GoblExecStep:
		fmt.Printf("%s: Exec %s\n", g.Name, step.Args)
	case GoblRunTaskStep:
		return <-RunTask(step.TaskName)
	}
	return GoblResult{nil, nil}
}

func (g *GoblTask) compile() error {
	for len(g.channel) > 0 {
		select {
		case t := <-g.channel:
			switch t := t.(type) {
			case GoblWatchStep:
				// Add to our watchers!
				matches, err := filepath.Glob(t.Path)
				if err != nil {
					fmt.Println(err)
				}
				g.watchPaths = append(g.watchPaths, matches...)
			case GoblExecStep:
				g.steps = append(g.steps, t)
			case GoblRunTaskStep:
				g.steps = append(g.steps, t)
			case GoblCatchTaskStep:
				g.steps = append(g.steps, t)
			case GoblResultTaskStep:
				g.steps = append(g.steps, t)
			}
		}
	}

	for _, file := range g.watchPaths {
		if err := g.watcher.Add(file); err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func (g *GoblTask) runLoop(resultChan chan GoblResult) {
	for {
		select {
		case shouldExit := <-g.runChannel:
			result := g.runSteps()
			if shouldExit == true {
				resultChan <- result
				return
			}
		case err := <-g.stopChannel:
			resultChan <- GoblResult{nil, err}
			return
		}
	}
}

func (g *GoblTask) watchLoop() {
	if len(g.watcher.WatchedFiles()) > 0 {
		fmt.Printf("👀 Watching: \n")
		for k, _ := range g.watcher.WatchedFiles() {
			fmt.Printf("\t%s\n", k)
		}
		// Watch events goroutine.
		go func() {
			g.runChannel <- false // Initial run
			for {
				select {
				case <-g.watcher.Event:
					g.runChannel <- false
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

func (g *GoblTask) run() chan GoblResult {
	result := make(chan GoblResult)

	go g.runLoop(result)

	go g.watchLoop()
	return result
}
