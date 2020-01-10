package main

import (
	"fmt"

	. "github.com/kettek/gobl/gobl"
)

func main() {
	task := Task("watch")
	task <- Watch("gobl/*")
	task <- Run("build")
	task <- Run("run")
	task <- Catch(func(err error) error {
		return nil
	})

	task2 := Task("build")
	task2 <- Exec("ls")
	task2 <- Catch(func(err error) error {
		// Ignore error
		return nil
	})
	task2 <- Result(func(r interface{}) {
		fmt.Println(r)
	})

	task3 := Task("run")
	task3 <- Exec("./client")

	task4 := Task("embeddedTest")
	task4 <- Exec("whoami")
	task4 <- Result(func(r interface{}) {
		fmt.Println("who am I:", r)
	})

	Go()
}
