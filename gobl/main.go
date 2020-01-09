package gobl

import "log"

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
type GoblRunStep struct {
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

func Run(taskName string) GoblStep {
	return GoblRunStep{
		TaskName: taskName,
	}
}

func Exec(arg string) GoblExecStep {
	return GoblExecStep{
		Args: []string{arg},
	}
}

func Go() {
	for _, v := range goblTasks {
		RunTask(v)
		return
	}
}

func RunTask(g GoblTask) error {
	for len(g.channel) > 0 {
		select {
		case t := <-g.channel:
			switch t := t.(type) {
			case GoblWatchStep:
				// Add to our watchers!
				g.steps = append(g.steps, t)
				log.Printf("Add watch %s\n", t.Path)
			case GoblExecStep:
				g.steps = append(g.steps, t)
				log.Printf("Exec %+v\n", t.Args)
			case GoblRunStep:
				g.steps = append(g.steps, t)
				log.Printf("Run %s\n", t.TaskName)
			}
		}
	}

	// IF we have watchers, then we'll spawn a go coroutine here!
	for _, step := range g.steps {
		switch step := step.(type) {
		case GoblWatchStep:
			log.Println("Watch")
		case GoblExecStep:
			log.Println("Exec")
		case GoblRunStep:
			run, ok := goblTasks[step.TaskName]
			if ok {
				err := RunTask(run)
				if err != nil {
					log.Fatal(err)
				}
			}
			log.Println("Run")
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
