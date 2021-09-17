package gobl

import (
	"fmt"
	"os"

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
		channel:     make(chan GoblStep, 99),
		stopChannel: make(chan error),
		runChannel:  make(chan bool),
		watcher:     watcher.New(),
	}
	return goblTasks[name]
}

// Watch sets up watching one or more glob paths.
func Watch(paths ...string) GoblStep {
	return GoblWatchStep{
		Paths: paths,
	}
}

// Catch handles any errors of preceding steps.
func Catch(f func(error) error) GoblStep {
	return GoblCatchTaskStep{
		Func: f,
	}
}

// Result handles the result of the previous step.
func Result(f func(interface{})) GoblStep {
	return GoblResultTaskStep{
		Func: f,
	}
}

// Run runs a given task by its name.
func Run(taskName string) GoblStep {
	return GoblRunTaskStep{
		TaskName: taskName,
	}
}

// Exec runs a command.
func Exec(args ...string) GoblExecStep {
	return GoblExecStep{
		Args: args,
	}
}

// Chdir changes the current directory to the one provided.
func Chdir(path string) GoblChdirStep {
	return GoblChdirStep{
		Path: path,
	}
}

// Exists checks if a path exists, returning an fs.FileInfo.
func Exists(path string) GoblExistsStep {
	return GoblExistsStep{
		Path: path,
	}
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
