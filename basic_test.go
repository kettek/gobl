package main

import (
	"testing"

	. "github.com/kettek/gobl/gobl"
)

func TestBasic(t *testing.T) {
	/*task := gobl.Task("Test")
	task.Watch <- "testdir/**"
	task.Exec <- "go build client"
	task.Catch <- func(err error) error {
		return nil
	}
	task.Run()*/

	task := Task("Test")
	task <- Watch("testdir/")
	task <- Run("Test 2")
	/*task <- gobl.Exec("go build client")
	task <- gobl.Catch(func(err error) error {
		return nil
	})*/

	task2 := Task("Test 2")
	task2 <- Exec("go build client")

	Go()
	//gobl.RunTask(task)
}
