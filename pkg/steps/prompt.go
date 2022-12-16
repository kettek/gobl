package steps

import (
	"fmt"
	"strings"
)

// PromptStep handles printing passed arguments or the results of the previous step if no arguments are passed.
type PromptStep struct {
	Message string
}

// Run prints the contents of the print step.
func (s PromptStep) Run(r Result) chan Result {
	result := make(chan Result)

	go func() {
		isYes := false
		for true {
			if len(s.Message) == 0 {
				fmt.Print(r.Result, " (y/n) ")
			} else {
				fmt.Print(s.Message, " (y/n) ")
			}
			var in string
			if _, err := fmt.Scan(&in); err != nil {
				result <- Result{Error: err}
				return
			}
			if strings.HasPrefix(strings.ToLower(in), "y") {
				isYes = true
				break
			} else if strings.HasPrefix(strings.ToLower(in), "n") {
				break
			}
		}
		result <- Result{Result: isYes}
	}()

	return result
}

// EndStep represents the end of a prompt.
type EndStep struct {
}

// Run just returns an empty result.
func (s EndStep) Run(r Result) chan Result {
	result := make(chan Result)

	go func() {
		result <- Result{}
	}()

	return result
}

// YesStep represents the beginning of a yes response to a prompt.
type YesStep struct {
}

// Run returns true in the result.
func (s YesStep) Run(r Result) chan Result {
	result := make(chan Result)

	go func() {
		result <- Result{Result: true}
	}()

	return result
}

// NoStep represents the beginning of a no response to a prompt.
type NoStep struct {
}

// Run returns false in the result.
func (s NoStep) Run(r Result) chan Result {
	result := make(chan Result)

	go func() {
		result <- Result{Result: false}
	}()

	return result
}
