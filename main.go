package gobl

import (
	"fmt"
	"os"
	"time"

	"github.com/kettek/gobl/pkg/colors"
	"github.com/kettek/gobl/pkg/messages"
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
	fmt.Printf("%s%s%s\n", colors.Info, messages.AvailableTasks, colors.Clear)
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
			fmt.Printf(messages.MissingTask+"\n", taskName)
			errChan <- steps.Result{Result: nil, Error: fmt.Errorf(messages.MissingTask, taskName), Context: nil}
		}()
	} else {
		fmt.Printf(messages.StartingTask+"\n", colors.Notice, colors.Clear, g.Name)
		t1 := time.Now()
		go func() {
			result := <-g.Execute()
			diff := time.Now().Sub(t1)

			if result.Result != nil {
				fmt.Printf("\t%s%v%s\n", colors.Info, result.Result, colors.Clear)
			}

			if result.Error != nil {
				fmt.Printf(messages.FailedTask+"\n", colors.Error, g.Name, colors.Clear, result.Error)
			} else {
				fmt.Printf(messages.CompletedTask+"\n", colors.Success, g.Name, diff, colors.Clear)
			}
			errChan <- result
		}()
	}
	return
}
