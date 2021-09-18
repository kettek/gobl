package steps

import "time"

// SleepStep causes a delay.
type SleepStep struct {
	Duration string
}

// Run sleeps for the given delay.
func (s SleepStep) Run(pr Result) chan Result {
	result := make(chan Result)

	go func() {
		d, err := time.ParseDuration(s.Duration)
		if err != nil {
			result <- Result{nil, err, nil}
			return
		}
		time.Sleep(d)
		result <- Result{s.Duration, nil, nil}
	}()

	return result
}
