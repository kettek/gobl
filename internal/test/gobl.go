package main

import (
	"fmt"

	. "github.com/kettek/gobl"
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
		Chdir("../chimera-rpg/go-meta/").
		Exec("./bin/client.exe").
		Chdir("../../gobl")

	Task("embeddedTest").
		Exec("whoami").
		Result(func(r interface{}) {
			fmt.Println("who am I:", r)
		})

	Task("chdirTest").
		Exists("gobl2").
		Catch(func(err error) error {
			return nil
		}).
		Chdir("gobl").
		Sleep("2s").
		Exec("git", "status").
		Chdir("..").
		//Exec("cmd", "/c", "pwd").
		Result(func(r interface{}) {
			fmt.Println("okay: ", r)
		})

	Task("envTest").
		Env("HUNGRY=true").
		Exec("env")

	Go()
}
