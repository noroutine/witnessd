// FSA - finite-state automatae, time bound
package fsa

import (
    "time"
)

// Transition function given input and state calculates new state for FSA
type TransitionFunc func(int, int) int

// Timeout function given state produces two values:
//  1. the time channel that ticks when FSA shall trigger timeout for given state
//  2. the transition function that calculates the next state for FSA in case of timeout
type TimeoutFunc func(int) (<-chan time.Time, func(int) int)

// terminate function given state determines if the FSA should stop execution
type TerminateFunc func(int) bool

type FSA struct  {
    Result chan int
    state int
    timeout TimeoutFunc 
    input chan int          // input events
    exec TransitionFunc
    end TerminateFunc
    term chan bool
}

// channel that is never written to
var neverTicks <-chan time.Time = make(chan time.Time, 0)

// Predefined function for building FSAs that never terminate automatically
func NeverTerminates() TerminateFunc {
    return func(s int) bool {
        return false
    }
}

// Predefined function for buildings FSAs that terminate on reaching one of states
func TerminatesOn(states ...int) TerminateFunc {
    var tss map[int]bool = make(map[int]bool, len(states))
    for _, ts := range states {
        tss[ts] = true
    }

    return func(s int) bool {
        _, ok := tss[s]
        return ok
    }
}

// Predefined timeout function for FSAs that never timeout in any state
func NeverTimesOut() TimeoutFunc {
    return func(int) (<-chan time.Time, func(int) int) {
        return neverTicks, func(s int) int {
            return s
        }
    }
}

// Create new FSA with given transition, terminate and timeout functions
func New(e TransitionFunc, end TerminateFunc, tout TimeoutFunc) (a *FSA) {    
    a = &FSA{
        Result: make(chan int),
        state: 0,
        input: make(chan int),
        exec: e,
        timeout: tout,
        end: end,
        term: make(chan bool),
    }
    go a.run()
    return
}

func (a *FSA) run() {
    for {
        timeTick, tfunc := a.timeout(a.state)
        select {
        case event := <- a.input: a.state = a.exec(a.state, event)
        case <- timeTick: a.state = tfunc(a.state)
        case <- a.term:
            a.Result <- a.state
            close(a.Result)
            close(a.input)
            close(a.term)
            return
        }

        if a.end(a.state) {
            go a.Terminate()
        }
    }
}

// Send the input to FSA
func (a *FSA) Send(input int) {
    a.input <- input
}

// Terminates FSA
func (a *FSA) Terminate() {
    a.term <- true
}