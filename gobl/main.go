package gobl

import (
	"fmt"
	"os"
	"time"

	"github.com/radovskyb/watcher"
)

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

// Task is a container for various steps.
func Task(name string) *GoblTask {
	goblTasks[name] = &GoblTask{
		Name:        name,
		stopChannel: make(chan error),
		runChannel:  make(chan bool),
		watcher:     watcher.New(),
	}
	return goblTasks[name]
}

func printInfo() {
	fmt.Printf("%s%s%s\n", InfoColor, "✨  Available Tasks", Clear)
	for k := range goblTasks {
		fmt.Printf("\t%s\n", k)
	}
}

// Go runs a specified task or lists all tasks if no task is specified.
func Go() {
	if len(os.Args) < 2 {
		printInfo()
		return
	}
	<-RunTask(os.Args[1])
}

// RunTask begins running a specifc named task.
func RunTask(taskName string) (errChan chan GoblResult) {
	g, ok := goblTasks[taskName]
	errChan = make(chan GoblResult)
	if !ok {
		go func() {
			fmt.Printf("🛑  task \"%s\" does not exist", taskName)
			errChan <- GoblResult{nil, fmt.Errorf("🛑 task \"%s\" does not exist", taskName), nil}
		}()
	} else {
		fmt.Printf("⚡  %sStarting Task%s \"%s\"\n", NoticeColor, Clear, g.Name)
		//g.compile()
		t1 := time.Now()
		go func() {
			goblResult := <-g.run()
			diff := time.Now().Sub(t1)

			if goblResult.Result != nil {
				fmt.Printf("\t%s%v%s\n", InfoColor, goblResult.Result, Clear)
			}

			if goblResult.Error != nil {
				fmt.Printf("❌  %sTask \"%s\" Failed%s: %s\n", ErrorColor, g.Name, Clear, goblResult.Error)
			} else {
				fmt.Printf("✔️  %sTask \"%s\" Complete in %s%s\n", SuccessColor, g.Name, diff, Clear)
			}
			errChan <- goblResult
		}()
	}
	return
}
