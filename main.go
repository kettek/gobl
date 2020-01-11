package main

import (
	"fmt"

	. "github.com/kettek/gobl/gobl"
)

func main() {
	Task("watch").
		Watch("gobl/*").
		Run("build").
		Run("run").
		Catch(func(err error) error {
			return nil
		})

	Task("build").
		Exec("ls").
		Catch(func(err error) error {
			// Ignore error
			return nil
		}).
		Result(func(r interface{}) {
			fmt.Println(r)
		})

	Task("run").
		Exec("./client")

	Task("embeddedTest").
		Exec("whoami").
		Result(func(r interface{}) {
			fmt.Println("who am I:", r)
		})

	Go()
}
