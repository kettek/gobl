package main

import (
	"testing"

	"github.com/kettek/gobl/gobl"
)

func TestBasic(t *testing.T) {
	/*task := gobl.Task("Test")
	task.Watch <- "testdir/**"
	task.Exec <- "go build client"
	task.Catch <- func(err error) error {
		return nil
	}
	task.Run()*/

	task := gobl.Task("Test")
	task <- gobl.Watch("testdir/")
	task <- gobl.Run("Test 2")
	/*task <- gobl.Exec("go build client")
	task <- gobl.Catch(func(err error) error {
		return nil
	})*/

	task2 := gobl.Task("Test 2")
	task2 <- gobl.Exec("go build client")

	gobl.Go()
	//gobl.RunTask(task)
}
