package gobl

import (
	"fmt"
	"os"
	"time"

	"github.com/kettek/gobl/pkg/globals"
	"github.com/kettek/gobl/pkg/steps"
	"github.com/kettek/gobl/pkg/task"
)

// Task is a container for various steps.
func Task(name string) *task.Task {
	task.Tasks[name] = task.NewTask(name, &Context{})
	return task.Tasks[name]
}

// PrintTasks prints the currently available tasks.
func PrintTasks() {
	fmt.Printf("%s%s%s\n", globals.InfoColor, globals.AvailableTasksMessage, globals.Clear)
	for k := range task.Tasks {
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
func RunTask(taskName string) (errChan chan steps.Result) {
	g, ok := task.Tasks[taskName]
	errChan = make(chan steps.Result)
	if !ok {
		go func() {
			fmt.Printf(globals.MissingTaskMessage+"\n", taskName)
			errChan <- steps.Result{Result: nil, Error: fmt.Errorf(globals.MissingTaskMessage, taskName), Context: nil}
		}()
	} else {
		fmt.Printf(globals.StartingTaskMessage+"\n", globals.NoticeColor, globals.Clear, g.Name)
		t1 := time.Now()
		go func() {
			result := <-g.Execute()
			diff := time.Now().Sub(t1)

			if result.Result != nil {
				fmt.Printf("\t%s%v%s\n", globals.InfoColor, result.Result, globals.Clear)
			}

			if result.Error != nil {
				fmt.Printf(globals.FailedTaskMessage+"\n", globals.ErrorColor, g.Name, globals.Clear, result.Error)
			} else {
				fmt.Printf(globals.CompletedTaskMessage+"\n", globals.SuccessColor, g.Name, diff, globals.Clear)
			}
			errChan <- result
		}()
	}
	return
}
