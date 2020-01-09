package main

import (
	. "github.com/kettek/gobl/gobl"
)

func main() {
	/*task := gobl.Task("Test")
	task.Watch <- "testdir/**"
	task.Exec <- "go build client"
	task.Catch <- func(err error) error {
		return nil
	}
	task.Run()*/

	task := Task("watch")
	task <- Watch("testdir/")
	task <- RunTask("build")
	/*task <- gobl.Exec("go build client")
	task <- gobl.Catch(func(err error) error {
		return nil
	})*/

	task2 := Task("build")
	task2 <- Exec("go build client")

	Go()
	//gobl.RunTask(task)
}
