package utils

import (
	"runtime"
)

type WorkResult struct {
	e error
	w Runner
}

type WorkChain struct {
	abort   chan struct{}
	results chan *WorkResult
}

func (self *WorkChain) Release() {
	close(self.abort)
}

type Runner interface {
	Run() error
}

func NewWorkChain(runners []Runner) (ret *WorkChain) {

	ret = &WorkChain{}
	ret.abort = make(chan struct{})

	if len(runners) == 0 {
		return
	}

	// Spawn as many workers as allowed threads
	workers := runtime.GOMAXPROCS(0)
	if len(runners) < workers {
		workers = len(runners)
	}

	// Create a task channel and spawn the verifiers
	var (
		inputs = make(chan int)
		done   = make(chan int, workers)
		errors = make([]*WorkResult, len(runners))
	)
	for i := 0; i < workers; i++ {
		go func() {
			for index := range inputs {
				runner := runners[index]
				errors[index] = &WorkResult{runner.Run(), runner}
				done <- index
			}
		}()
	}

	ret.results = make(chan *WorkResult, len(runners))
	go func() {
		defer close(inputs)
		var (
			in, out = 0, 0
			checked = make([]bool, len(runners))
			inputs  = inputs
		)
		for {
			select {
			case inputs <- in:
				if in++; in == len(runners) {
					// Reached end of headers. Stop sending to workers.
					inputs = nil
				}
			case index := <-done:
				for checked[index] = true; checked[out]; out++ {
					ret.results <- errors[out]
					if out == len(runners)-1 {
						return
					}
				}
			case <-ret.abort:
				return
			}
		}
	}()
	return
}
