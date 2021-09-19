package task

import (
	"fmt"

	"github.com/kettek/gobl/pkg/messages"
)

// Tasks is our global task name to *Task map.
var Tasks []*Task

// AddTask adds a task by its name. If a task with the same name already exists, it is replaced.
func AddTask(t *Task) {
	index := getTaskIndex(t.Name)
	if index != -1 {
		fmt.Printf(messages.ExistingTask+"\n", t.Name)
		Tasks[index] = t
	} else {
		Tasks = append(Tasks, t)
	}
}

// GetTask returns the Task matching the provided name.
func GetTask(name string) *Task {
	for _, t := range Tasks {
		if t.Name == name {
			return t
		}
	}
	return nil
}

func getTaskIndex(name string) int {
	for i, t := range Tasks {
		if t.Name == name {
			return i
		}
	}
	return -1
}
