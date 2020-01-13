package gobl

import (
	"fmt"
	"os"

	"github.com/radovskyb/watcher"
)

var (
	InfoColor    = Purple
	NoticeColor  = Teal
	WarnColor    = Yellow
	ErrorColor   = Red
	SuccessColor = Green
)

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

func Task(name string) *GoblTask {
	goblTasks[name] = &GoblTask{
		Name:        name,
		channel:     make(chan GoblStep, 99),
		stopChannel: make(chan error),
		runChannel:  make(chan bool),
		watcher:     watcher.New(),
	}
	return goblTasks[name]
}

func Watch(path string) GoblStep {
	return GoblWatchStep{
		Path: path,
	}
}

func Catch(f func(error) error) GoblStep {
	return GoblCatchTaskStep{
		Func: f,
	}
}

func Result(f func(interface{})) GoblStep {
	return GoblResultTaskStep{
		Func: f,
	}
}

func Run(taskName string) GoblStep {
	return GoblRunTaskStep{
		TaskName: taskName,
	}
}

func Exec(args ...string) GoblExecStep {
	return GoblExecStep{
		Args: args,
	}
}

func printInfo() {
	fmt.Printf("%s%s%s\n", InfoColor, "✨  Available Tasks", Clear)
	for k, _ := range goblTasks {
		fmt.Printf("\t%s\n", k)
	}
}

func Go() {
	if len(os.Args) < 2 {
		printInfo()
		return
	}
	<-RunTask(os.Args[1])
}

func RunTask(taskName string) (errChan chan GoblResult) {
	g, ok := goblTasks[taskName]
	errChan = make(chan GoblResult)
	if !ok {
		go func() {
			fmt.Printf("🛑  task \"%s\" does not exist", taskName)
			errChan <- GoblResult{nil, fmt.Errorf("🛑 task \"%s\" does not exist", taskName)}
		}()
	} else {
		fmt.Printf("⚡  %sStarting Task%s \"%s\"\n", NoticeColor, Clear, g.Name)
		//g.compile()
		go func() {
			goblResult := <-g.run()

			if goblResult.Result != nil {
				fmt.Printf("\t%s%v%s\n", InfoColor, goblResult.Result, Clear)
			}

			if goblResult.Error != nil {
				fmt.Printf("❌  %sTask \"%s\" Failed%s: %s\n", ErrorColor, g.Name, Clear, goblResult.Error)
			} else {
				fmt.Printf("✔️  %sTask \"%s\" Complete%s\n", SuccessColor, g.Name, Clear)
			}
			errChan <- goblResult
		}()
	}
	return
}
