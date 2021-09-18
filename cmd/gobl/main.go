/*
The gobl command simply calls "go run gobl.go ..." in the working directory.
*/
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	var p = "gobl.go"

	// Ensure our file exists and is not a directory.
	info, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("missing %s\n", p)
		} else {
			fmt.Printf("%s\n", err)
		}
		os.Exit(1)
	}
	if info.IsDir() {
		fmt.Printf("%s is a directory\n", p)
		os.Exit(1)
	}

	// Set up our command and run it.
	args := []string{"run", p}
	if len(os.Args) > 1 {
		args = append(args, os.Args[1:]...)
	}
	cmd := exec.Command("go", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}
