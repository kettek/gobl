package steps

import (
	"fmt"
	"strings"
	"sync"
)

// ParallelStep runs tasks in parallel.
type ParallelStep struct {
	TaskNames []string
}

type parallelOperation struct {
	name       string
	runChannel chan Result
	result     Result
}

// Run uses wait groups.
func (s ParallelStep) Run(r Result) chan Result {
	var wg sync.WaitGroup
	parallelResult := make(chan Result)
	var parallelOperations []*parallelOperation

	for _, t := range s.TaskNames {
		taskResult := &parallelOperation{
			name:       t,
			runChannel: r.Context.RunTask(t),
		}
		parallelOperations = append(parallelOperations, taskResult)
		wg.Add(1)
		go func(pOp *parallelOperation) {
			defer wg.Done()
			pOp.result = <-pOp.runChannel
		}(taskResult)
	}
	go func() {
		wg.Wait()
		var err error
		var errStrings []string
		for _, pr := range parallelOperations {
			if pr.result.Error != nil {
				errStrings = append(errStrings, fmt.Sprintf("%s -> %s", pr.name, pr.result.Error))
			}
		}
		if len(errStrings) > 0 {
			err = fmt.Errorf(strings.Join(errStrings, ","))
		}
		parallelResult <- Result{
			Result:  parallelOperations,
			Error:   err,
			Context: r.Context,
		}
	}()
	return parallelResult
}
