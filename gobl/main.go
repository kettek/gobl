package gobl

import (
	"fmt"
	"github.com/radovskyb/watcher"
	"os"
)

func Task(name string) chan GoblStep {
	goblTasks[name] = &GoblTask{
		Name:        name,
		channel:     make(chan GoblStep, 99),
		stopChannel: make(chan error),
		runChannel:  make(chan bool),
		watcher:     watcher.New(),
	}
	return goblTasks[name].channel
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
	fmt.Printf("Available Tasks (run with \"%s %s\")\n", os.Args[0], "MyTask")
	for k, _ := range goblTasks {
		fmt.Printf("\t%s\n", k)
	}
}

func Go() {
	if len(os.Args) < 2 {
		printInfo()
		return
	}
	if goblResult := <-RunTask(os.Args[1]); goblResult.Error != nil {
		if goblResult.Result != nil {
			fmt.Println(goblResult.Result)
		}
		if goblResult.Error != nil {
			fmt.Println(goblResult.Error)
		}
	}
}

func RunTask(taskName string) (errChan chan GoblResult) {
	g, ok := goblTasks[taskName]
	if !ok {
		errChan = make(chan GoblResult)
		go func() {
			errChan <- GoblResult{nil, fmt.Errorf("task \"%s\" does not exist", taskName)}
		}()
	} else {
		g.compile()
		errChan = g.run()
	}
	return
}
