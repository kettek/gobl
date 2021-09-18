package gobl

import (
	"fmt"
	"os"
	"time"

	"github.com/radovskyb/watcher"
)

// Tasks is our global task name to *GoblTask map.
var Tasks = make(map[string]*GoblTask)

// Our name to color map.
var (
	InfoColor    = Purple
	NoticeColor  = Teal
	WarnColor    = Yellow
	ErrorColor   = Red
	SuccessColor = Green
)

// Our colors to escape codes map.
var (
	Black   = "\033[1;30m"
	Red     = "\033[1;31m"
	Green   = "\033[1;32m"
	Yellow  = "\033[1;33m"
	Purple  = "\033[1;34m"
	Magenta = "\033[1;35m"
	Teal    = "\033[1;36m"
	White   = "\033[1;37m"
	Clear   = "\033[0m"
)

// Our messages.
var (
	AvailableTasksMessage = "✨  Available Tasks"
	MissingTaskMessage    = "🛑  task \"%s\" does not exist"
	StartingTaskMessage   = "⚡  %sStarting Task%s \"%s\""
	CompletedTaskMessage  = "✔️  %sTask \"%s\" Complete in %s%s"
	FailedTaskMessage     = "❌  %sTask \"%s\" Failed%s: %s"
)

// Result represents the result of a step.
type Result struct {
	Result interface{}
	Error  error
	Task   *GoblTask // TODO: Move Task to some sort of GoblContext that gets passed into steps.
}

// Task is a container for various steps.
func Task(name string) *GoblTask {
	Tasks[name] = &GoblTask{
		Name:        name,
		stopChannel: make(chan error),
		runChannel:  make(chan bool),
		watcher:     watcher.New(),
	}
	return Tasks[name]
}

// PrintTasks prints the currently available tasks.
func PrintTasks() {
	fmt.Printf("%s%s%s\n", InfoColor, AvailableTasksMessage, Clear)
	for k := range Tasks {
		fmt.Printf("\t%s\n", k)
	}
}

// Go runs a specified task or lists all tasks if no task is specified.
func Go() {
	if len(os.Args) < 2 {
		PrintTasks()
		return
	}
	<-RunTask(os.Args[1])
}

// RunTask begins running a specifc named task.
func RunTask(taskName string) (errChan chan Result) {
	g, ok := Tasks[taskName]
	errChan = make(chan Result)
	if !ok {
		go func() {
			fmt.Printf(MissingTaskMessage+"\n", taskName)
			errChan <- Result{nil, fmt.Errorf(MissingTaskMessage, taskName), nil}
		}()
	} else {
		fmt.Printf(StartingTaskMessage+"\n", NoticeColor, Clear, g.Name)
		t1 := time.Now()
		go func() {
			result := <-g.run()
			diff := time.Now().Sub(t1)

			if result.Result != nil {
				fmt.Printf("\t%s%v%s\n", InfoColor, result.Result, Clear)
			}

			if result.Error != nil {
				fmt.Printf(FailedTaskMessage+"\n", ErrorColor, g.Name, Clear, result.Error)
			} else {
				fmt.Printf(CompletedTaskMessage+"\n", SuccessColor, g.Name, diff, Clear)
			}
			errChan <- result
		}()
	}
	return
}
