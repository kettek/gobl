package gobl

import (
	"fmt"
	"os"
)

var goblTasks = make(map[string]GoblTask)

type GoblTask struct {
	Name     string
	watching []string
	running  []GoblTask
	steps    []GoblStep
	channel  chan GoblStep
}

type GoblStep interface {
}
type GoblWatchStep struct {
	Path string
}
type GoblRunTaskStep struct {
	TaskName string
}
type GoblExecStep struct {
	Args []string
}
type GoblCanKill struct {
}

func Task(name string) chan GoblStep {
	goblTasks[name] = GoblTask{
		Name:    name,
		channel: make(chan GoblStep, 99),
	}
	return goblTasks[name].channel
}

func Watch(path string) GoblStep {
	return GoblWatchStep{
		Path: path,
	}
}

func RunTask(taskName string) GoblStep {
	return GoblRunTaskStep{
		TaskName: taskName,
	}
}

func Exec(arg string) GoblExecStep {
	return GoblExecStep{
		Args: []string{arg},
	}
}

func printInfo() {
	fmt.Printf("Available Tasks (run with \"%s %s\")", os.Args[0], "MyTask")
	for k, _ := range goblTasks {
		fmt.Printf("\t%s\n", k)
	}
}

func Go() {
	if len(os.Args) < 2 {
		printInfo()
		return
	}
	if task, ok := goblTasks[os.Args[1]]; ok {
		runTask(task)
		return
	}
	fmt.Println("No such task exists.")
}

func runTask(g GoblTask) error {
	for len(g.channel) > 0 {
		select {
		case t := <-g.channel:
			switch t := t.(type) {
			case GoblWatchStep:
				// Add to our watchers!
				//g.steps = append(g.steps, t)
				fmt.Printf("Add watch %s\n", t.Path)
			case GoblExecStep:
				g.steps = append(g.steps, t)
				fmt.Printf("Exec %+v\n", t.Args)
			case GoblRunTaskStep:
				g.steps = append(g.steps, t)
				fmt.Printf("Add Run %s\n", t.TaskName)
			}
		}
	}

	// IF we have watchers, then we'll spawn a go coroutine here!
	for _, step := range g.steps {
		switch step := step.(type) {
		case GoblExecStep:
			fmt.Println("Exec")
		case GoblRunTaskStep:
			run, ok := goblTasks[step.TaskName]
			if ok {
				err := runTask(run)
				if err != nil {
					fmt.Println(err)
				}
			}
			fmt.Println("Ran subtask")
		}
	}

	return nil
	/*if len(g.watching) > 0 {
		for watch := range g.Watch {
			g.watching = append(g.watching, watch)
		}
	}
	if len(g.running) > 0 {
		for run := range g.Run {
			fmt.Printf("added run: %+v\n", run)
		}
	}

	fmt.Printf("%+v\n", g)*/
}
